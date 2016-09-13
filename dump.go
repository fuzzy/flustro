package main

import (
	"fmt"
	"path"
	"reflect"
	"time"

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

func dumpStruct(s interface{}, h bool) {
	typ := reflect.TypeOf(s)
	switch typ.Kind() {
	case reflect.Struct:
		// First, lets find our longest element
		le, te := longestElement(s)
		// Now let's print our header
		if h {
			output := ""
			for i := 0; i < typ.NumField(); i++ {
				p := typ.Field(i)
				if !p.Anonymous {
					switch p.Type.Kind() {
					case reflect.String, reflect.Uint32, reflect.Float64, reflect.Int:
						if i < (te - 1) {
							output = Strappend(output, fmt.Sprintf("%s%s| ", p.Name, buffer((le[0]+1), len(p.Name))))
						} else {
							output = Strappend(output, fmt.Sprintf("%s%s", p.Name, buffer((le[0]+1), len(p.Name))))
						}
					}
				}
			}
			Info <- output
		}
		// Now let's dump our contents
		output := ""
		for i := 0; i < typ.NumField(); i++ {
			p := typ.Field(i)
			if !p.Anonymous {
				switch p.Type.Kind() {
				case reflect.String, reflect.Uint32, reflect.Float64, reflect.Int:
					value := fmt.Sprintf("%v", reflect.ValueOf(s).Field(i))
					if i < (te - 1) {
						output = Strappend(output, fmt.Sprintf("%s%s| ", value, buffer((le[0]+1), len(value))))
					} else {
						output = Strappend(output, fmt.Sprintf("%s%s", value, buffer((le[0]+1), len(value))))
						Info <- output
						output = ""
					}
				}
			}
		}

	}
}

func DumpWhisperFile(c *cli.Context) error {
	for _, f := range c.Args() {
		data := Metadata{}
		db, err := whisper.Open(f)
		if err != nil {
			Error <- fmt.Sprintln(err)
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
					Error <- fmt.Sprintln(e.Error())
					return e
				}
				for _, point := range p {
					data.Archives[i].Points = append(data.Archives[i].Points,
						Point{
							Timestamp: point.Timestamp,
							Value:     point.Value,
							Time:      point.Time(),
						})
				}
			}
		}
		dumpStruct(data, true)
		time.Sleep(200 * time.Millisecond)
		fmt.Println("")
		for i, v := range data.Archives {
			if i == 0 || c.Bool("P") {
				dumpStruct(v, true)
			} else {
				dumpStruct(v, false)
			}
			if c.Bool("P") {
				if c.Int("A") == 40960 || c.Int("A") == i {
					for n, p := range v.Points {
						if n == 0 {
							dumpStruct(p, true)
						} else {
							dumpStruct(p, false)
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
		Aliases:     []string{"du"},
		Usage:       "Dump metadata, and optionally data of a whisper file",
		Description: "Dump the metadata, such as retention periods, and point data from a given whisper file",
		ArgsUsage:   "<whisperFile>",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "P", Usage: "Dump data points"},
			cli.IntFlag{Name: "A", Usage: "Archive # to dump. (Requires -P)", Value: 40960},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          DumpWhisperFile,
	})
}
