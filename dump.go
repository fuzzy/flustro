package main

import (
	"fmt"
	"log"

	. "github.com/fuzzy/gocolor"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

func DumpWhisperFile(c *cli.Context) error {
	for _, f := range c.Args() {
		db, err := whisper.Open(f)
		if err != nil {
			fmt.Println("Could not open whisper file:", err)
			return err
		} else {
			defer db.Close()
			// Dump header information
			am := db.Header.Metadata.AggregationMethod
			mr := db.Header.Metadata.MaxRetention
			ac := db.Header.Metadata.ArchiveCount
			if !c.Bool("c") {
				fmt.Printf("\n%s: %s\n",
					String("File").Yellow().Bold(), f)
				fmt.Printf("%s: %s, %s: %d, %s: %d\n\n",
					String("Agg Method").Yellow().Bold(), am,
					String("Max Retention").Yellow().Bold(), mr,
					String("Archives").Yellow().Bold(), ac)
			} else {
				fmt.Printf("\nFile: %s\n", f)
				fmt.Printf("Agg Method: %s, Max Retention: %d, Archives: %d\n\n", am, mr, ac)
			}
			// Now show the archive headers
			if !c.Bool("c") {
				fmt.Printf("%s   | %s    | %s    | %s  | %s | %s\n",
					String("Archive").Yellow().Bold(),
					String("Offset").Yellow().Bold(),
					String("Points").Yellow().Bold(),
					String("Interval").Yellow().Bold(),
					String("Retention").Yellow().Bold(),
					String("Size").Yellow().Bold())
			} else {
				fmt.Println("Archive   | Offset    | Points    | Interval  | Retention | Size")
			}
			for i, a := range db.Header.Archives {
				fmt.Printf("%-10s| %-10s| %-10s| %-10s| %-10s| %-10s\n",
					fmt.Sprint(i),
					fmt.Sprint(a.Offset),
					fmt.Sprint(a.Points),
					fmt.Sprint(a.SecondsPerPoint),
					fmt.Sprint(a.Retention()),
					fmt.Sprint(a.Size()))
				// TODO: Add in how many points are filled, and how many aren't yadda yadda
				// And finally the data points if desired.
				if c.Bool("P") {
					// dumpArchives(db)
					for i := range db.Header.Archives {
						p, e := db.DumpArchive(i)
						if e != nil {
							fmt.Printf("%s: Failed to read archive: %s\n", String("ERROR").Red().Bold(), e)
							return e
						}
						for n, point := range p {
							if point.Timestamp > 0 && point.Value > 0 {
								if !c.Bool("c") {
									fmt.Printf("%s: %d, %10.35g\n",
										String(fmt.Sprintf("%d", n)).Yellow(),
										point.Timestamp, point.Value)
								} else {
									fmt.Printf("%d: %d, %10.35g\n", n, point.Timestamp, point.Value)
								}
							}
						}
					}
				}
			}
		}
	}
	fmt.Println("")
	return nil
}

func init() {
	Commands = append(Commands, cli.Command{
		Name:        "dump",
		Aliases:     []string{"d"},
		Usage:       "Dump metadata, and optionally data of a whisper file",
		Description: "Dump the metadata, such as retention periods, and point data from a given whisper file",
		ArgsUsage:   "<whisperFile>",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "P", Usage: "Dump data points"},
			cli.BoolFlag{Name: "c", Usage: "Disable colors in output"},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          DumpWhisperFile,
	})
}
