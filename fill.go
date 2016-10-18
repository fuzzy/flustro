package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

var (
	fillJobs int
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

	// now for each archive in src, if there's one that matches in dst, let's backfill
	for sIdx, sArc := range sDb.Header.Archives {
		for dIdx, dArc := range dDb.Header.Archives {
			// if we hit this we have found a compatible archive
			if sArc.Points == dArc.Points && sArc.SecondsPerPoint == dArc.SecondsPerPoint {
				sp, se := sDb.DumpArchive(sIdx)
				if se != nil {
					return se
				}
				dp, de := dDb.DumpArchive(dIdx)
				if de != nil {
					return de
				}
				// for every point in the src archive
				for sidx := 0; sidx < len(sp); sidx++ {
					// check all the points in the dst archive
					for didx := 0; didx < len(dp); didx++ {
						// if any dst point has the same timestamp
						if sp[sidx].Timestamp == dp[didx].Timestamp {
							// and the values aren't the same
							if dp[didx].Value != sp[sidx].Value {
								// update the dst
								dp[didx].Value = sp[sidx].Value
							}
						}
					}
				}
			}
		}
	}

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
		if c.Int("j") > maxFills {
			maxFills = c.Int("j")
		}
		// if both our arguments are files
		if isFile(srcArg) && isFile(dstArg) {
			NewInfo(fmt.Sprintf("Backfill: %s -> %s", srcArg, dstArg))
			backFill(srcArg, dstArg)
		} else if isDir(srcArg) && isDir(dstArg) {
			return
		} else {
			fmt.Println("You cannot mix files and directories as src and dst.")
			os.Exit(1)
		}
	}
	return
}

/*

func BackFill(c *cli.Context) {
	if len(c.Args()) < 2 {
		Error <- "Invalid arguments. See 'flustro help fill' for more information."
		os.Exit(1)
	} else {
		// declare our variables
		var srcObj Dirstate
		var dstObj Dirstate
		srcDir := c.Args().Get(0)
		dstDir := c.Args().Get(1)

		// Now let's do the heavy lifting
		if isDir(srcDir) && isDir(dstDir) {
			// First let's get our dir contents
			srcObj = ListDir(srcDir)
			dstObj = ListDir(dstDir)
			// then spawn our worker pool, and get to processing
			for i := 0; i < c.Int("j"); i++ {
				go Filler()
			}
			// next we'll start processing through our srcObj and dstObj lists and
			// backfill everything that's present in both locations
			overlap := Overlap{
				Source:      srcObj.Location,
				Destination: dstObj.Location,
				Contents:    make(map[string][]string),
			}
			overlap_c := 0
			for k, _ := range srcObj.Contents {
				if _, ok := dstObj.Contents[k]; ok {
					for _, v := range srcObj.Contents[k] {
						for _, dv := range dstObj.Contents[k] {
							if v == dv {
								if _, ok := overlap.Contents[k]; ok {
									overlap.Contents[k] = append(overlap.Contents[k], v)
									overlap_c++
								} else {
									overlap.Contents[k] = []string{v}
									overlap_c++
								}
							}
						}
					}
				}
			}
			Info <- fmt.Sprintf("%d entries shared between %s and %s.",
				overlap_c,
				overlap.Source,
				overlap.Destination)
			for k, v := range overlap.Contents {
				for _, f := range v {
					DataChan <- map[string]string{
						"Source":      fmt.Sprintf("%s/%s/%s", overlap.Source, k, f),
						"Destination": fmt.Sprintf("%s/%s/%s", overlap.Destination, k, f),
					}
				}
			}
		} else if isFile(srcDir) && isFile(dstDir) {
			// we only need one worker for this job
			go Filler()
			DataChan <- map[string]string{
				"Source":      srcDir,
				"Destination": dstDir,
			}
		} else {
			Error <- fmt.Sprintf("SRC and DST must be either both files or both dirs.")
		}

		time.Sleep(500 * time.Millisecond)
	}
	return
}

*/
func init() {
	fillJobs = 0
	maxFills = runtime.GOMAXPROCS(0) * 2
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
				Value: runtime.GOMAXPROCS(0) * 2,
			},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          BackFill,
	})
}
