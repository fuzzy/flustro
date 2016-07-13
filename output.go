package main

import (
	"fmt"
	"reflect"
	"sort"

	. "github.com/fuzzy/gocolor"
)

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

func DumpStruct(s interface{}, h bool) {
	typ := reflect.TypeOf(s)
	switch typ.Kind() {
	case reflect.Struct:
		// First, lets find our longest element
		le, te := longestElement(s)
		// Now let's print our header
		if h {
			for i := 0; i < typ.NumField(); i++ {
				p := typ.Field(i)
				if !p.Anonymous {
					// fmt.Printf("%s %s: %s %v\n", indent(), p.Name, p.Type, reflect.ValueOf(s).Field(i))
					switch p.Type.Kind() {
					case reflect.String, reflect.Uint32, reflect.Float64, reflect.Int:
						if i < (te - 1) {
							fmt.Printf("%s%s| ", String(p.Name).Cyan().Bold(), buffer((le[0]+1), len(p.Name)))
						} else {
							fmt.Printf("%s%s", String(p.Name).Cyan().Bold(), buffer((le[0]+1), len(p.Name)))
						}
					}
				}
			}
			fmt.Println("")
		}
		// Now let's dump our contents
		for i := 0; i < typ.NumField(); i++ {
			p := typ.Field(i)
			if !p.Anonymous {
				switch p.Type.Kind() {
				case reflect.String, reflect.Uint32, reflect.Float64, reflect.Int:
					value := fmt.Sprintf("%v", reflect.ValueOf(s).Field(i))
					if i < (te - 1) {
						fmt.Printf("%s%s| ", value, buffer((le[0]+1), len(value)))
					} else {
						fmt.Printf("%s%s\n", value, buffer((le[0]+1), len(value)))
					}
				}
			}
		}

	}
}
