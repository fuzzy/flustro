package main

import (
	"log"
	"os"

	. "github.com/fuzzy/gocolor"
)

var (
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func init() {
	Debug = log.New(os.Stdout,
		string(String("DEBUG ").Cyan().Bold()),
		log.Lshortfile)
	Info = log.New(os.Stdout,
		string(String("Info ").Green().Bold()),
		0)
	Warn = log.New(os.Stderr,
		string(String("Warning ").Yellow().Bold()),
		0)
	Error = log.New(os.Stderr,
		string(String("Error ").Red().Bold()),
		0)
}
