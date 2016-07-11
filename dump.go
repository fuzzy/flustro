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
				fmt.Printf("\n%s: %s, %s: %d, %s: %d\n\n",
					String("Agg Method").Yellow(), am,
					String("Max Retention").Yellow(), mr,
					String("Archives").Yellow(), ac)
			} else {
				fmt.Printf("\nAgg Method: %s, Max Retention: %d, Archives: %d\n\n", am, mr, ac)
			}
			// Now show the archive headers
			for i, a := range db.Header.Archives {
				if !c.Bool("c") {
					fmt.Printf("%s #%d, %s: %d, %s: %d\n",
						String("Archive").Yellow(), i,
						String("Offset").Yellow(), a.Offset,
						String("Points").Yellow(), a.Points)
					fmt.Printf("%s: %d, %s: %d, %s: %d\n\n",
						String("Interval").Yellow(), a.SecondsPerPoint,
						String("Retention").Yellow(), a.Retention(),
						String("Size").Yellow(), a.Size())
				} else {
					fmt.Printf("Archive #%d, Offset: %d, Points: %d\n", i, a.Offset, a.Points)
					fmt.Printf("Interval: %d, Retention: %d, Size: %d\n\n", a.SecondsPerPoint, a.Retention(), a.Size())
				}
				// TODO: Add in how many points are filled, and how many aren't yadda yadda
				// And finally the data points if desired.
				if c.Bool("P") {
					// dumpArchives(db)
					for i := range db.Header.Archives {
						p, e := db.DumpArchive(i)
						if e != nil {
							log.Fatalln("Failed to read archive:", e)
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
	return nil
}

func init() {
	Commands = append(Commands, cli.Command{
		Name:        "dump",
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
