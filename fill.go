package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	. "github.com/fuzzy/gocolor"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

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
		fmt.Printf("%s: The files have a mismatched set of archives.\n",
			String("ERROR").Red().Bold())
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
					fmt.Println("WE ARE GO, I REPEAT, WE ARE FUCKING GO!")
					sp, se := sDb.DumpArchive(i)
					if se != nil {
						fmt.Printf("%s: %s\n", String("ERROR").Red().Bold(), se)
						os.Exit(1)
					}
					dp, de := dDb.DumpArchive(i)
					if de != nil {
						fmt.Printf("%s: %s\n", String("ERROR").Red().Bold(), de)
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
			e := fmt.Sprintf("%s: Dir comparison not complete yet",
				String("WARNING").Yellow().Bold())
			fmt.Println(e)
			// Let's get this dir walking in there then shall we?
			srcFiles := []string{}
			cwd, _ := os.Getwd()
			os.Chdir(args[0])
			filepath.Walk(".", func(p string, i os.FileInfo, e error) error {
				chkErr(e)
				if !i.IsDir() {
					srcFiles = append(srcFiles, p)
				}
				return nil
			})
			os.Chdir(args[1])
			fmt.Println(len(srcFiles), "files in", args[0])
			// now for every file we have, let's see if we have a match in dst
			for _, v := range srcFiles {
				if _, err := os.Stat(v); err == nil {
					fmt.Printf("Backfill: %s -> %s\n",
						fmt.Sprintf("%s%s", args[0], v),
						fmt.Sprintf("%s%s", args[1], v))
				}
			}
			os.Chdir(cwd)
		} else {
			if !backfillFile(args[0], args[1]) {
				e := fmt.Sprintf("%s: There has been an error.",
					String("ERROR").Red().Bold())
				fmt.Println(e)
				return errors.New(e)
			}
		}
	} else {
		var e string
		if !c.Bool("c") {
			e = fmt.Sprintf("%s: Wrong number of paramters given.\n",
				String("ERROR").Red().Bold())
			e = fmt.Sprintf("%s%s: Try '%s help fill' for more information",
				e,
				String("ERROR").Red().Bold(),
				path.Base(os.Args[0]))
		} else {
			e = fmt.Sprintf("ERROR: Wrong number of parameters given.\n")
			e = fmt.Sprintf("%sERROR: Try '%s help fill' for more information.",
				e,
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
