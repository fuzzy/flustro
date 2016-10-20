package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/kisielk/whisper-go/whisper"
	"github.com/splatpm/subhuman"
	"github.com/urfave/cli"
)

var (
	fillJobs int
	doneJobs int
	maxFills int
	fillLock *sync.Mutex
)

type Dirstate struct {
	Location string
	Contents map[string][]string
}

type Overlap struct {
	Source      string
	Destination string
	Contents    map[string][]string
}

func backFill(src string, dst string) error {
	// get the lock on our jobs tracker
	fillLock.Lock()
	// increment the tracker
	fillJobs++
	// and release the lock
	fillLock.Unlock()

	// open our file handles
	sDb, sErr := whisper.Open(src)
	dDb, dErr := whisper.Open(dst)
	// check our error status
	if sErr != nil {
		return sErr
	} else if dErr != nil {
		return dErr
	}
	// defer db closes
	defer sDb.Close()
	defer dDb.Close()
	if sDb.Header.Metadata.ArchiveCount != dDb.Header.Metadata.ArchiveCount {
		return fmt.Errorf("The files have a mismatched set of archives.")
	} else {
		// Now we'll start processing the archives, checking as we go to see if they
		// are matched. That way we at least fill in what we can, possibly....
		for i, a := range sDb.Header.Archives {
			// The offset
			if a.Offset == dDb.Header.Archives[i].Offset {
				// and the number of points
				if a.Points == dDb.Header.Archives[i].Points {
					// and finally the interval
					if a.SecondsPerPoint == dDb.Header.Archives[i].SecondsPerPoint {
						// ok, now let's get rolling through the archives
						sp, se := sDb.DumpArchive(i)
						if se != nil {
							return se
						}
						dp, de := dDb.DumpArchive(i)
						if de != nil {
							return de
						}
						for idx := 0; idx < len(sp); idx++ {
							if sp[idx].Timestamp != 0 && sp[idx].Value != 0 {
								if dp[idx].Timestamp == 0 || dp[idx].Value == 0 {
									dp[idx].Timestamp = sp[idx].Timestamp
									dp[idx].Value = sp[idx].Value
									dDb.Update(dp[idx])
								}
							}
						}
					}
				}
			}
		}
	}
	fillLock.Lock()
	fillJobs--
	doneJobs++
	fillLock.Unlock()
	// and finally let the caller know nothing went wrong.
	return nil
}

func BackFill(c *cli.Context) {
	if len(c.Args()) < 2 {
		fmt.Println("You fucked up, look at the help output.")
	} else {
		srcArg := c.Args().Get(0)
		dstArg := c.Args().Get(1)
		// if we requested more jobs than the default, make sure we update that
		if c.Int("j") != maxFills {
			maxFills = c.Int("j")
		}
		// if both our arguments are files
		if isFile(srcArg) && isFile(dstArg) {
			// NewInfo(fmt.Sprintf("Backfill: %s -> %s", srcArg, dstArg))
			StartTime := time.Now().Unix()
			backFill(srcArg, dstArg)
			NewInfo(fmt.Sprintf("Filled %d of 1 files: %s", doneJobs, subhuman.HumanTimeColon(time.Now().Unix()-StartTime)))
		} else if isDir(srcArg) && isDir(dstArg) {
			// get our directory lists
			overlap, totFills := CollateDirs(srcArg, dstArg)
			Status(fmt.Sprintf("Filled %d of %d files (%3.02f%%): %s",
				doneJobs,
				totFills,
				((float32(doneJobs) / float32(totFills)) * 100.0),
				subhuman.HumanTimeColon(time.Now().Unix()-StartTime)))
			StartTime := time.Now().Unix()
			for k, v := range overlap.Contents {
				for _, f := range v {
					// this waits until there are free job slots
					for fillJobs >= maxFills {
						now := time.Now().Unix()
						rate := int64(float64(doneJobs) / (float64(now) - float64(StartTime)))
						Status(
							fmt.Sprintf("Filled %d of %d files (%3.02f%%): %s @ %d/sec eta: %s",
								doneJobs,
								totFills,
								((float32(doneJobs) / float32(totFills)) * 100.0),
								subhuman.HumanTimeColon(now-StartTime),
								rate,
								subhuman.HumanTimeColon(int64(totFills-doneJobs)/rate)))
						time.Sleep(200 * time.Millisecond)
					}
					go backFill(fmt.Sprintf("%s/%s/%s",
						overlap.Source, k, f),
						fmt.Sprintf("%s/%s/%s",
							overlap.Destination, k, f))
					time.Sleep(100 * time.Millisecond)
				}
			}
			// wait for any stray jobs to finish
			for doneJobs < totFills {
				time.Sleep(10 * time.Millisecond)
			}
			// and print the final status
			Status(fmt.Sprintf("Filled %d of %d files (%3.02f%%): %s",
				doneJobs,
				totFills,
				((float32(doneJobs) / float32(totFills)) * 100.0),
				subhuman.HumanTimeColon(time.Now().Unix()-StartTime)))
			fmt.Println("")
		} else {
			fmt.Println("You cannot mix files and directories as src and dst.")
			os.Exit(1)
		}
	}
	return
}

func init() {
	fillJobs = 0
	doneJobs = 0
	maxFills = runtime.GOMAXPROCS(0) - 2
	fillLock = &sync.Mutex{}
	Commands = append(Commands, cli.Command{
		Name:        "fill",
		Aliases:     []string{"f"},
		Usage:       "Backfill datapoints in the dst from the src",
		Description: "Backfill datapoints in the dst from the src",
		ArgsUsage:   "<src(File|Dir)> <dst(File|Dir)>",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "j",
				Usage: "Number of workers (for directory recursion)",
				Value: runtime.GOMAXPROCS(0) - 2,
			},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          BackFill,
	})
}
