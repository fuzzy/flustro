package main

import (
	"fmt"
	"os"

	. "github.com/fuzzy/gocolor"
)

func Debug(s string) {
	fmt.Println(String("DEBUG").Cyan().Bold(), s)
}

func Info(s string) {
	fmt.Println(String("INFO").Green().Bold(), s)
}

func Warn(s string) {
	fmt.Println(String("WARNING").Yellow().Bold(), s)
}

func Error(s string) {
	fmt.Println(String("ERROR").Red().Bold(), s)
}

func Fatal(s string) {
	fmt.Println(String("FATAL").Red().Bold(), s)
	os.Exit(1)
}
