package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fuzzy/gocolor"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

type FillState struct {
	StartTime int64
	Count     int64
	CountLock sync.Mutex
	CountChan chan bool
	SrcTotal  int64
	DstTotal  int64
	Overlap   int64
}

func (f FillState) Increment() {
	f.CountLock.Lock()
	f.Count++
	f.CountChan <- true
	f.CountLock.Unlock()
}

func (f FillState) DumpState() {
	// grab our lock, this ensures accurate reporting, and gives a bit more
	// of a concurrency throttle. This is not necessarily a bad thing.
	f.CountLock.Lock()
	r := (time.Now().Unix() - f.StartTime)
	c := f.Count
	if r > 0 {
		c = (f.Count / r)
	}
	p := (float64(f.Count) / float64(f.Overlap)) * 100.00
	fmt.Printf("%s %-6s/%-6s (%6.02f%%) in %d seconds @ %-6s/sec",
		gocolor.String("Info").Green().Bold(),
		humanize.Comma(f.Count),
		humanize.Comma(f.Overlap),
		p, HumanTime(int(r)), c)
	f.CountLock.Unlock()
}

var (
	State FillState
)

func stateWorker(t chan bool) {
	for {
		select {
		case sig := <-State.CountChan:
			if sig {
				State.DumpState()
			} else {
				return
			}
		case <-t:
			return
		}
	}
}

func fillWorker(d chan map[string]string, t chan bool) {
	for {
		msg := <-d
		if msg["SRC"] == "__EXIT__" && msg["DST"] == "__EXIT__" {
			return
		} else {
			if isFile(msg["SRC"]) && isFile(msg["DST"]) {
				if !backfillFile(msg["SRC"], msg["DST"]) {
					Error.Printf("S:%s D:%s - Backfill operation failed.",
						msg["SRC"],
						msg["DST"])
					os.Exit(1)
				} else {
					State.Increment()
				}
			}
		}
	}
}

func backfillFile(s string, d string) bool {
	// Open our filehandles
	sDb, sErr := whisper.Open(s)
	dDb, dErr := whisper.Open(d)
	if !chkErr(sErr) || !chkErr(dErr) {
		os.Exit(1)
	}
	// Defer their closings
	defer sDb.Close()
	defer dDb.Close()

	// Now for a series of checks, first to ensure that both
	// files have the same number of archives in them.
	if sDb.Header.Metadata.ArchiveCount != dDb.Header.Metadata.ArchiveCount {
		Error.Println("The files have a mismatched set of archives.")
		return false
	}

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
						Error.Println(se.Error())
						os.Exit(1)
					}
					dp, de := dDb.DumpArchive(i)
					if de != nil {
						Error.Println(de.Error())
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
	return true
}

func Filler(c *cli.Context) error {
	if len(c.Args()) == 2 {
		args := c.Args()
		if isDir(args[0]) && isDir(args[1]) {
			State.Count = 0

			// First things first, let's spawn our pool of workers.
			dataCh := make(chan map[string]string, 1)
			timeout := make(chan bool, 30)
			go stateWorker(timeout)
			for i := 0; i < c.Int("j"); i++ {
				go fillWorker(dataCh, timeout)
			}

			// Let's get this dir walking in there then shall we?
			srcFiles := listFiles(args[0])
			dstFiles := listFiles(args[1])

			// Now let's find all our overlap
			overlap := []string{}
			for a := 0; a < len(srcFiles); a++ {
				for b := 0; b < len(dstFiles); b++ {
					if srcFiles[a] == dstFiles[b] {
						overlap = append(overlap, srcFiles[a])
					}
				}
			}

			// And display some stats
			Info.Printf("srcDir: %d files, dstDir: %d files, %d overlap.",
				len(srcFiles),
				len(dstFiles),
				len(overlap))
			State.Overlap = int64(len(overlap))
			State.SrcTotal = int64(len(srcFiles))
			State.DstTotal = int64(len(dstFiles))

			// Now we can push in all our data, and let our workers do their
			// lovely little thing. Ahhhhh concurrency.
			StartTime = time.Now().Unix()
			for i := 0; i < len(overlap); i++ {
				dataCh <- map[string]string{
					"SRC": fmt.Sprintf("%s/%s", args[0], overlap[i]),
					"DST": fmt.Sprintf("%s/%s", args[1], overlap[i]),
				}
			}

			// And while they're off doing that, here we will just sit and watch
			// the return channel and count things. Once we have all our
			// backfill operations accounted for, we can reap all of our workers
			// and carry on.
			for State.Count < int64(len(overlap)) {
				time.Sleep(1 * time.Second)
			}
			// And finally let's reap all our children
			for idx := 0; idx < c.Int("j"); idx++ {
				dataCh <- map[string]string{
					"SRC": "__EXIT__",
					"DST": "__EXIT__",
				}
			}

			fmt.Println("")
			Info.Println((State.Overlap / (time.Now().Unix() - StartTime)),
				"whisper files processed per second.")
		} else {
			if !backfillFile(args[0], args[1]) {
				e := fmt.Sprintf("Error while backfilling.")
				Error.Println(e)
				return errors.New(e)
			}
		}
	} else {
		var e string
		e = fmt.Sprintf("Wrong number of paramters given.")
		Error.Println(e)
		Error.Printf("Try '%s help fill' for more information",
			path.Base(os.Args[0]))
		return errors.New(e)
	}
	return nil
}

func init() {
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
				Value: runtime.GOMAXPROCS(0),
			},
			cli.BoolFlag{Name: "c", Usage: "Prevent colors from being used"},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          Filler,
	})
}
