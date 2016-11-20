package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"git.thwap.org/splat/gout"
	"github.com/kisielk/whisper-go/whisper"
	"github.com/urfave/cli"
)

var (
	fillJobs int
	doneJobs int
	maxFills int
	fillLock *sync.Mutex
)

// BEGIN: Points sort.Interface implimentation

type Points []whisper.Point

func (p Points) Len() int {
	return len(p)
}

func (p Points) Less(i, j int) bool {
	return p[i].Timestamp < p[j].Timestamp
}

func (p Points) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func NewPoints(p []whisper.Point) Points {
	retv := Points{}
	for i := 0; i < len(p); i++ {
		if p[i].Value != 0 {
			retv[i] = p[i]
		}
	}
	return retv
}

// END: Points sort.Interface implimentation

type Dirstate struct {
	Location string
	Contents map[string][]string
}

type Overlap struct {
	Source      string
	Destination string
	Contents    map[string][]string
}

func fill(src, dst string) error {
	// open our archives
	sdb, serr := whisper.Open(src)
	ddb, derr := whisper.Open(dst)
	// and error check
	if serr != nil {
		return serr
	} else if derr != nil {
		return derr
	}
	defer sdb.Close()
	defer ddb.Close()
	// find the oldest point in time
	stm := time.Now().Unix() - int64(sdb.Header.Metadata.MaxRetention)
	dtm := time.Now().Unix() - int64(ddb.Header.Metadata.MaxRetention)
	// and process the archives
	for _, a := range sdb.Header.Archives {
		// let's setup the time boundaries
		from := time.Now().Unix() - int64(a.Retention())
		// grab the src and dest data and error check
		_, sp, se := sdb.FetchUntil(uint32(from), uint32(stm))
		if se != nil {
			return se
		}
		_, dp, de := ddb.FetchUntil(uint32(from), uint32(dtm))
		if de != nil {
			return de
		}
		// Migrate our []whisper.Point to a Points type for sorting
		spts := NewPoints(sp)
		dpts := NewPoints(dp)
		// and sort that bitch
		sort.Sort(spts)
		sort.Sort(dpts)
		pts := Points{}
		// now gather an array of points that are non-null and who's corresponding
		// element in the destination archive is not identical
		for _, spnt := range spts {
			for _, dpnt := range dpts {
				if spnt.Value <= 0 {
					if spnt.Timestamp == dpnt.Timestamp {
						if spnt.Value != dpnt.Value {
							pts = append(pts, spnt)
						}
					}
				}
			}
		}
		ddb.UpdateMany(pts)
	}
	// and send the all clear
	return nil
}

func fillArchives(c *cli.Context) {
	if len(c.Args()) < 2 {
		gout.Error("Invalid arguments")
	}
	src := c.Args().Get(0)
	dst := c.Args().Get(1)
	if c.Int("j") != maxFills {
		maxFills = c.Int("j")
	}
	st_time := time.Now().Unix()
	if isFile(src) && isFile(dst) {
		fill(src, dst)
		st_time = time.Now().Unix()
		gout.Info("This file took %s", gout.HumanTimeConcise(time.Now().Unix()-st_time))
	} else if isDir(src) && isDir(dst) {
		ovr, fills := CollateDirs(src, dst)
		st_time = time.Now().Unix()
		for k, v := range ovr.Contents {
			for _, f := range v {
				// hold off if we need to
				for fillJobs >= maxFills {
					time.Sleep(100 * time.Millisecond)
				}
				go func() {
					fillLock.Lock()
					fillJobs++
					fillLock.Unlock()
					s := fmt.Sprintf("%s/%s/%s", ovr.Source, k, f)
					d := fmt.Sprintf("%s/%s/%s", ovr.Destination, k, f)
					fill(s, d)
					fillLock.Lock()
					fillJobs--
					doneJobs++
					fillLock.Unlock()
				}()
				rn_time := (time.Now().Unix() - st_time)
				done := int(fills - doneJobs)
				if rn_time == 0 {
					rn_time = 1
				}
				if done == 0 {
					done = 1
				}
				speed := int(int64(doneJobs) / rn_time)
				remain := (fills - doneJobs)
				var rm_time int64
				if remain > 0 && speed > 0 {
					rm_time = int64(remain / speed)
				} else {
					rm_time = 1
				}
				cons := gout.ConsInfo()
				strl := len(fmt.Sprintf("Speed: %d f/sec, Elapsed: %s, Eta: %s %3d%%",
					speed,
					gout.HumanTimeColon(rn_time),
					gout.HumanTimeColon(rm_time),
					int((float32(doneJobs)/float32(fills))*100.0))) + 7
				prgl := (cons.Col - uint16(strl))
				gout.Status("Speed: %d f/sec, Elapsed: %s, Eta: %s, %3d%% %s",
					speed,
					gout.HumanTimeColon(rn_time),
					gout.HumanTimeColon(rm_time),
					int((float32(doneJobs)/float32(fills))*100.0),
					gout.Progress(int(prgl), int((float32(doneJobs)/float32(fills))*100.0)))
			}
		}
		// wait for all jobs to finish
		for doneJobs < fills {
			time.Sleep(100 * time.Millisecond)
		}
	}
	cons := gout.ConsInfo()
	tot_time := time.Now().Unix() - st_time
	strl := len(fmt.Sprintf("Speed: %d f/sec, Elapsed: %s, Eta: %s, %3d%%",
		int(int64(doneJobs)/tot_time),
		gout.HumanTimeColon(tot_time),
		gout.HumanTimeColon(int64(0)),
		100)) + 7
	prgl := (cons.Col - uint16(strl))
	gout.Info("Speed: %d f/sec, Elapsed: %s, Eta: %s, %3d%% %s",
		int(int64(doneJobs)/tot_time),
		gout.HumanTimeColon(tot_time),
		gout.HumanTimeColon(int64(0)),
		100,
		gout.Progress(int(prgl), 100))
}

func init() {
	gout.Setup(true, false, true, "")
	gout.Output.Prompts["status"] = fmt.Sprintf("%s%s%s",
		gout.String(".").Cyan(),
		gout.String(".").Bold().Cyan(),
		gout.String(".").Bold().White())
	gout.Output.Prompts["info"] = gout.Output.Prompts["status"]
	gout.Output.Prompts["debug"] = fmt.Sprintf("%s.%s.%s",
		gout.String(".").Purple(),
		gout.String(".").Bold().Purple(),
		gout.String(".").Bold().White())
	fillJobs = 0
	doneJobs = 0
	maxFills = runtime.GOMAXPROCS(0) - 2
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
				Value: runtime.GOMAXPROCS(0) - 2,
			},
		},
		SkipFlagParsing: false,
		HideHelp:        false,
		Hidden:          false,
		Action:          fillArchives,
	})
}
