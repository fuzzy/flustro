package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"git.thwap.org/splat/gout"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

var mutex = &sync.Mutex{}
var running = 0

func replayThisFile(fpath string, mname string, sock net.Conn) {
	mutex.Lock()
	running++
	mutex.Unlock()
	if isFile(fpath) {
		db, de := whisper.Open(fpath)
		if de != nil {
			gout.Error(de.Error())
			panic(de.Error())
		}
		points, pe := db.DumpArchive(0)
		if pe != nil {
			gout.Error(pe.Error())
			panic(pe.Error())
		}
		for _, point := range points {
			if point.Value > 0 && point.Timestamp > 0 {
				fmt.Fprintf(sock, "%s %f %d\n", mname, point.Value, point.Timestamp)
			}
		}
		db.Close()
	}
	mutex.Lock()
	running--
	mutex.Unlock()
}

func Replay(c *cli.Context) error {
	sock, se := net.Dial("tcp", fmt.Sprintf("%s:%d", c.String("H"), c.Int("P")))
	if se != nil {
		gout.Error(se.Error())
		panic(se.Error())
	}
	for _, arg := range c.Args() {
		fpath := fmt.Sprintf("%s/%s", c.String("R"), arg)
		if isFile(fpath) {
			mname := fmt.Sprintf("%s.%s", strings.Replace(path.Dir(arg), "/", ".", -1), strings.Replace(path.Base(arg), ".wsp", "", -1))
			replayThisFile(fpath, mname, sock)
		} else if isDir(fpath) {
			data := ListDir(fpath)
			datac := data.Count()
			datai := 0
			for dirn, dirc := range data.Contents {
				for _, fn := range dirc {
					fpath := fmt.Sprintf("%s/%s/%s/%s", c.String("R"), arg, dirn, fn)
					if isFile(fpath) {
						mname := fmt.Sprintf("%s.%s", strings.Replace(arg, "/", ".", -1), strings.Replace(path.Base(fpath), ".wsp", "", -1))
						for running >= runtime.GOMAXPROCS(0) {
							time.Sleep(250 * time.Millisecond)
						}
						gout.Status("%d of %d: %s", datai, datac, mname)
						replayThisFile(fpath, mname, sock)
						datai++
					}
				}
			}
		}
	}
	sock.Close()
	fmt.Println("")
	return nil
}

func init() {
	this_dir, _ := os.Getwd()
	Commands = append(Commands, cli.Command{
		Name:        "replay",
		Aliases:     []string{"re"},
		Usage:       "Replay metrics to a graphite endpoint.",
		Description: "Replay the metrics from a given file or directory, to a given carbon endpoint.\n   Only the first archive will be replayed at this point.",
		ArgsUsage:   "<whisperFile/Dir>",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "H", Usage: "Specify the remote host (default localhost)", Value: "localhost"},
			cli.IntFlag{Name: "P", Usage: "Port number (default 2003)", Value: 2003},
			cli.StringFlag{Name: "R", Usage: "Root directory (for metric name generation)", Value: this_dir},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          Replay,
	})
}
