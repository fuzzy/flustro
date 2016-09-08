package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

var (
	DataChan  chan map[string]string
	FillerSig chan bool
	Count     int32
	Mutex     *sync.Mutex
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

func Filler() {
	for {
		select {
		case msg := <-DataChan:
			Debug <- fmt.Sprintf("Backfill: %s -> %s", msg["Source"], msg["Destination"])

			// Open our filehandles
			sDb, sErr := whisper.Open(msg["Source"])
			dDb, dErr := whisper.Open(msg["Destination"])
			if !chkErr(sErr) {
				Error <- fmt.Sprintf("%s: %s", msg["Source"], sErr.Error())
			} else if !chkErr(dErr) {
				Error <- fmt.Sprintf("%s: %s", msg["Destination"], dErr.Error())
				os.Exit(1)
			}
			// Defer their closings
			defer sDb.Close()
			defer dDb.Close()

			// Now for a series of checks, first to ensure that both
			// files have the same number of archives in them.
			if sDb.Header.Metadata.ArchiveCount != dDb.Header.Metadata.ArchiveCount {
				Error <- fmt.Sprintln("The files have a mismatched set of archives.")
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
									Error <- fmt.Sprintf("%s: %s", msg["Source"], se.Error())
									os.Exit(1)
								}
								dp, de := dDb.DumpArchive(i)
								if de != nil {
									Error <- fmt.Sprintln(de.Error())
									os.Exit(1)
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
			Mutex.Lock()
			Count++
			Mutex.Unlock()
		case <-FillerSig:
			FillerSig <- true
			return
		}
	}
}

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
			time.Sleep(500 * time.Millisecond)
			start_t := int32(time.Now().Unix())
			for k, v := range overlap.Contents {
				for _, f := range v {
					DataChan <- map[string]string{
						"Source":      fmt.Sprintf("%s/%s/%s", overlap.Source, k, f),
						"Destination": fmt.Sprintf("%s/%s/%s", overlap.Destination, k, f),
					}
				}
			}
			for {
				runtime := (int32(time.Now().Unix()) - start_t)
				if len(DataChan) == 0 {
					time.Sleep(2 * time.Second)
					runrate := (float32(overlap_c) / float32(runtime))
					Info <- fmt.Sprintf("%d files processed in %d sec @ %.02f/sec.", overlap_c, runtime, runrate)
					time.Sleep(100 * time.Millisecond)
					return
				}
				runrate := (float32(Count) / float32(runtime))
				Progress <- fmt.Sprintf("%d files processed in %d sec @ %.02f/sec.", Count, runtime, runrate)
				time.Sleep(1 * time.Second)
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

func init() {
	DataChan = make(chan map[string]string, 8192)
	Mutex = &sync.Mutex{}
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
				Value: (runtime.GOMAXPROCS(0) * 2),
			},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          BackFill,
	})
}
