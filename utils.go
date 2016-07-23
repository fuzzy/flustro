package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
)

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
		Error.Println(e.Error())
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

func listFiles(p string) (retv []string) {
	cwd, _ := os.Getwd()
	os.Chdir(p)
	Info.Printf("Examinging directory: %s", p)
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

func HumanTime(s int) string {
	tdesc := map[string]int{
		"s": 1,
		"m": 60,
		"h": 60 * 60,
		"d": (60 * 60) * 24,
		"w": ((60 * 60) * 24) * 7,
		"y": ((60 * 60) * 24) * 365,
	}
	keys := []string{"y", "w", "d", "h", "m", "s"}
	retv := ""
	for _, t := range keys {
		val := (s / tdesc[t])
		tgt := (val * tdesc[t])
		if s >= tdesc[t] {
			retv = Strappend(retv, fmt.Sprintf("%02d%s", val, t))
			s = (s - tgt)
		} else {
			retv = Strappend(retv, fmt.Sprintf("00%s", t))
		}
	}
	return retv
}
