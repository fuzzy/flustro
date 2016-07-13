package main

import (
	"fmt"
	"path"
	"time"

	. "github.com/fuzzy/gocolor"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

type Point struct {
	Timestamp uint32
	Value     float64
	Time      time.Time
}

type Archive struct {
	Archive   int
	Offset    uint32
	NumPoints uint32
	Interval  uint32
	Retention uint32
	Size      uint32
	Points    []Point
}

type Metadata struct {
	File         string
	AggMethod    string
	MaxRetention uint32
	NumArchives  uint32
	Archives     []Archive
}

func DumpWhisperFile(c *cli.Context) error {
	data := Metadata{}
	for _, f := range c.Args() {
		db, err := whisper.Open(f)
		if err != nil {
			fmt.Println("Could not open whisper file:", err)
			return err
		} else {
			defer db.Close()
			// Record our metadata
			data.AggMethod = db.Header.Metadata.AggregationMethod.String()
			data.File = path.Base(f)
			data.MaxRetention = db.Header.Metadata.MaxRetention
			data.NumArchives = db.Header.Metadata.ArchiveCount
			// Now the archive metadata
			for i, a := range db.Header.Archives {
				data.Archives = append(data.Archives, Archive{
					Archive:   i,
					Offset:    a.Offset,
					NumPoints: a.Points,
					Interval:  a.SecondsPerPoint,
					Retention: a.Retention(),
					Size:      a.Size(),
					Points:    []Point{},
				})
				p, e := db.DumpArchive(i)
				if e != nil {
					fmt.Printf("%s: Failed to read archive: %s\n", String("ERROR").Red().Bold(), e)
					return e
				}
				for _, point := range p {
					data.Archives[i].Points = append(data.Archives[i].Points, Point{
						Timestamp: point.Timestamp,
						Value:     point.Value,
						Time:      point.Time(),
					})
				}
			}
		}
	}
	DumpStruct(data, true)
	fmt.Println("")
	for i, v := range data.Archives {
		if i == 0 {
			DumpStruct(v, true)
		} else {
			DumpStruct(v, false)
		}
	}
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
