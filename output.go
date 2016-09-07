package main

import (
	"fmt"
	"os"

	. "github.com/fuzzy/gcl"
)

var (
	Info      chan string
	Error     chan string
	Debug     chan string
	Progress  chan string
	OutputSig chan os.Signal
)

func Outputter() {
	for {
		select {
		case msg := <-Info:
			if !StripColor {
				fmt.Printf("%s %s\n", String(">>").Bold().Green(), msg)
			} else {
				fmt.Printf("I> %s\n", msg)
			}
		case msg := <-Error:
			if !StripColor {
				fmt.Printf("%s %s\n", String(">>").Bold().Red(), msg)
			} else {
				fmt.Printf("E> %s\n", msg)
			}
		case msg := <-Debug:
			if !StripColor {
				fmt.Printf("%s %s\n", String(">>").Bold().Cyan(), msg)
			} else {
				fmt.Printf("D> %s\n", msg)
			}
		case msg := <-Progress:
			ws := consInfo()
			buff := (int(ws.Col) - (len(msg) + 1))
			fmt.Printf("%s %s%s\r", String(">").Bold().Green(), msg, padding(buff))
		case <-OutputSig:
			fmt.Println("signal")
		}
	}
}

func init() {
	Info = make(chan string, 255)
	Error = make(chan string, 255)
	Debug = make(chan string, 255)
	Progress = make(chan string, 255)
	OutputSig = make(chan os.Signal, 1)
}
