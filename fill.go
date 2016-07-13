package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"

	. "github.com/fuzzy/gocolor"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

func isDir(p string) bool {
	f, e := os.Stat(p)
	if e != nil {
		return false
	}
	return f.IsDir()
}

func chkErr(e error) bool {
	if e != nil {
		fmt.Printf("%s: %s\n", String("ERROR").Red().Bold(), e)
		return false
	} else {
		return true
	}
}

func backfillFile(s string, d string) bool {
	// Open our filehandles
	srcDb, srcErr := whisper.Open(s)
	chkErr(srcErr)
	dstDb, dstErr := whisper.Open(d)
	chkErr(dstErr)
	// Defer their closings
	defer srcDb.Close()
	defer dstDb.Close()

	// Now for a series of checks, first to ensure that both
	// files have the same number of archives in them.
	if srcDb.Header.Metadata.ArchiveCount != dstDb.Header.Metadata.ArchiveCount {
		fmt.Printf("%s: The files have a mismatched set of archives.\n", String("ERROR").Red().Bold())
		return false
	}

	// Now we'll start processing the archives, checking as we go to see if they are matched.
	// that way we at least fill in what we can, possibly....
	for i, a := range srcDb.Header.Archives {
		// The offset
		if a.Offset == dstDb.Header.Archives[i].Offset {
			// and the number of points
			if a.Points == dstDb.Header.Archives[i].Points {
				// and finally the interval
				if a.SecondsPerPoint == dstDb.Header.Archives[i].SecondsPerPoint {
					// ok, now let's get rolling through the archives
					fmt.Println("WE ARE GO, I REPEAT, WE ARE FUCKING GO!")
					sp, se := srcDb.DumpArchive(i)
					if se != nil {
						fmt.Printf("%s: %s\n", String("ERROR").Red().Bold(), se)
						os.Exit(1)
					}
					dp, de := dstDb.DumpArchive(i)
					if de != nil {
						fmt.Printf("%s: %s\n", String("ERROR").Red().Bold(), de)
						os.Exit(1)
					}
					for idx := 0; idx < len(sp); idx++ {
						if sp[idx].Timestamp != 0 && sp[idx].Value != 0 {
							if dp[idx].Timestamp == 0 || dp[idx].Value == 0 {
								fmt.Printf("SRC: %s %d %f\n", sp[idx].Time(), sp[idx].Timestamp, sp[idx].Value)
								fmt.Printf("DST: %s %d %f\n", dp[idx].Time(), dp[idx].Timestamp, dp[idx].Value)
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
			e := fmt.Sprintf("%s: Dir comparison not complete yet", String("ERROR").Red().Bold())
			fmt.Println(e)
			return errors.New(e)
		} else {
			if !backfillFile(args[0], args[1]) {
				e := fmt.Sprintf("%s: There has been an error.", String("ERROR").Red().Bold())
				fmt.Println(e)
				return errors.New(e)
			}
		}
	} else {
		var e string
		if !c.Bool("c") {
			e = fmt.Sprintf("%s: Wrong number of paramters given.\n%s: Try '%s help fill' for more information",
				String("ERROR").Red().Bold(),
				String("ERROR").Red().Bold(),
				path.Base(os.Args[0]))
		} else {
			e = fmt.Sprintf("ERROR: Wrong number of parameters given.\nERROR: Try '%s help fill' for more information.",
				path.Base(os.Args[0]))
		}
		fmt.Println(e)
		return errors.New(e)
	}
	return nil
}

func init() {
	Commands = append(Commands, cli.Command{
		Name:        "fill",
		Aliases:     []string{"f"},
		Usage:       "Backfill datapoints in the dst(file|dir) from the src(file|dir)",
		Description: "Backfill datapoints in the dst(file|dir) from the src(file|dir)",
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
