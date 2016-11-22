package gout

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"

	eval "github.com/Knetic/govaluate"
)

var (
	Output  output
	Winsize winsize
	Logfile *os.File
)

// data structures

type output struct {
	Prompts   map[string]string
	Debug     bool
	Quiet     bool
	Verbose   bool
	ToFile    bool
	Throbber  []string
	lastThrob int
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// console info and utilities

func ConsInfo() winsize {
	ws := winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	if int(retCode) == -1 {
		panic(errno)
	}
	return ws
}

func repeat(c string, n int) string {
	retv := ""
	for i := 0; i <= n; i++ {
		retv += "#"
	}
	return retv
}

func padding(spaces int) string {
	pad := " "
	retv := ""
	for i := 0; i < spaces; i++ {
		retv = fmt.Sprintf("%s%s", retv, pad)
	}
	return retv
}

// Output functions
func consoleOutput(t string, e string, f string, args ...interface{}) {
	Winsize = ConsInfo()
	if Output.ToFile {
		log.Printf("%s %s\n", t, fmt.Sprintf(f, args...))
	}
	fmt.Printf("%s %s %s%s",
		Output.Prompts[t],
		fmt.Sprintf(f, args...),
		padding(int(Winsize.Col)-(7+len(fmt.Sprintf(f, args...)))),
		e)
}

func Info(f string, args ...interface{}) {
	if !Output.Quiet {
		consoleOutput("info", "\n", f, args...)
	}
}

func Debug(f string, args ...interface{}) {
	if Output.Debug {
		consoleOutput("debug", "\n", f, args...)
	}
}

func Warn(f string, args ...interface{}) {
	if Output.Verbose {
		consoleOutput("warn", "\n", f, args...)
	}
}

func Error(f string, args ...interface{}) {
	consoleOutput("error", "\n", f, args...)
}

func Status(f string, args ...interface{}) {
	consoleOutput("status", "\r", f, args...)
}

func Progress(l int, p int) string {
	rl := l - 2
	var rslt interface{}
	var sp string
	if p < 100 {
		expr, _ := eval.NewEvaluableExpression(fmt.Sprintf("%d * 0.%02d", rl, p))
		rslt, _ = expr.Evaluate(nil)
		sp = repeat("#", int(rslt.(float64)))
	} else {
		rslt = rl
		sp = repeat("#", rslt.(int))
	}
	pd := padding(rl - len(sp))
	return fmt.Sprintf("[%s%s]", sp, pd)
}

func Throbber() string {
	if Output.lastThrob == len(Output.Throbber)+1 {
		Output.lastThrob = 1
	} else {
		Output.lastThrob++
	}
	return Output.Throbber[Output.lastThrob-1]
}

// Output setup function
func Setup(d bool, q bool, v bool, f string) {
	if len(f) > 0 {
		Logfile, e := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if e != nil {
			panic(e)
		}
		log.SetOutput(Logfile)
		Output = output{
			Prompts:   make(map[string]string),
			Debug:     d,
			Quiet:     q,
			Verbose:   v,
			ToFile:    true,
			Throbber:  []string{},
			lastThrob: 0,
		}
	} else {
		Output = output{
			Prompts:   make(map[string]string),
			Debug:     d,
			Quiet:     q,
			Verbose:   v,
			ToFile:    false,
			Throbber:  []string{},
			lastThrob: 0,
		}
	}
	Output.Prompts["info"] = "INFO"
	Output.Prompts["warn"] = "WARN"
	Output.Prompts["debug"] = "DEBUG"
	Output.Prompts["error"] = "ERROR"
	Output.Prompts["status"] = ""
	Output.Throbber = []string{"-", "\\", "|", "/"}
}

// Setup example
/*
func init() {
	Winsize = consInfo()
	Setup(true, false, true, false)
	Output.Prompts["info"] = fmt.Sprintf("%s%s%s",
		String(".").Cyan(),
		String(".").Bold().Cyan(),
		String(".").Bold().White())
	Output.Prompts["warn"] = fmt.Sprintf("%s%s%s",
		String(".").Yellow(),
		String(".").Bold().Yellow(),
		String(".").Bold().White())
	Output.Prompts["debug"] = fmt.Sprintf("%s%s%s",
		String(".").Purple(),
		String(".").Bold().Purple(),
		String(".").Bold().White())
	Output.Prompts["error"] = fmt.Sprintf("%s%s%s",
		String(".").Red(),
		String(".").Bold().Red(),
		String(".").Bold().White())
	Output.Prompts["status"] = fmt.Sprintf("%s%s%s",
		String("-").Cyan(),
		String("-").Bold().Cyan(),
		String("-").Bold().White())
}
*/
