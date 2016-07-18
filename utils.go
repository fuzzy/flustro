// utils.go
package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"

	. "github.com/fuzzy/gocolor"
)

func isDir(p string) bool {
	f, e := os.Stat(p)
	if e != nil {
		return false
	}
	return f.IsDir()
}

func chkErr(e error) bool {
	if e != nil {
		fmt.Printf("%s: %s\n", String("ERROR").Red().Bold(), e)
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
