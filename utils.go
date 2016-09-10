package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"syscall"
	"unsafe"
)

// terminal width
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func consInfo() winsize {
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

func padding(spaces int) string {
	pad := " "
	retv := ""
	for i := 0; i < spaces; i++ {
		retv = fmt.Sprintf("%s%s", retv, pad)
	}
	return retv
}

// end terminal width

// TODO: CLEAN THIS SHIT UP

func isDir(p string) bool {
	f, e := os.Stat(p)
	if e != nil {
		return false
	}
	return f.IsDir()
}

func isFile(p string) bool {
	if _, err := os.Stat(p); err == nil && !isDir(p) {
		return true
	} else {
		return false
	}
}

func chkErr(e error) bool {
	if e != nil {
		Error <- fmt.Sprintln(e.Error())
		return false
	} else {
		return true
	}
}

var Indent int = 0

func indent() string {
	retv := ""
	for m := 0; m <= Indent; m++ {
		retv = fmt.Sprintf("%s|", retv)
	}
	return retv
}

func reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func longestElement(s interface{}) ([]int, int) {
	retv := []int{}
	totl := 0
	typ := reflect.TypeOf(s)
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		if !p.Anonymous {
			switch p.Type.Kind() {
			case reflect.String, reflect.Uint32, reflect.Float64, reflect.Int:
				retv = append(retv, len(p.Name))
				retv = append(retv, len(reflect.ValueOf(s).Field(i).String()))
				totl++
			}
		}
	}

	sort.Ints(retv)
	retv = reverse(retv)
	return retv, totl
}

func buffer(m int, h int) string {
	retv := ""
	for i := h; i < m; i++ {
		retv = fmt.Sprintf("%s ", retv)
	}
	return retv
}

func ListDir(dir string) Dirstate {
	retv := Dirstate{Location: dir, Contents: make(map[string][]string)}

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	filepath.Walk(".", func(p string, i os.FileInfo, e error) error {
		if e != nil {
			Error <- e.Error()
		} else {
			if !i.IsDir() {
				// let's seperate the components
				b := filepath.Dir(p)
				f := filepath.Base(p)
				// and analyze them
				if _, ok := retv.Contents[b]; ok {
					retv.Contents[b] = append(retv.Contents[b], f)
				} else {
					retv.Contents[b] = []string{f}
				}
			}
		}
		return nil
	})
	os.Chdir(cwd)

	return retv
}

func listFiles(p string) (retv []string) {
	cwd, _ := os.Getwd()
	os.Chdir(p)
	Info <- fmt.Sprintf("Examinging directory: %s", p)
	filepath.Walk(".", func(p string, i os.FileInfo, e error) error {
		chkErr(e)
		if !i.IsDir() {
			retv = append(retv, p)
		}
		return nil
	})
	os.Chdir(cwd)
	return retv
}

func Strappend(p string, a string) string {
	b := bytes.NewBufferString(p)
	for i := 0; i < len(a); i++ {
		b.WriteByte(a[i])
	}
	return b.String()
}
