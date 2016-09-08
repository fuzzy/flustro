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
		ws := consInfo()
		select {
		case msg := <-Info:
			buff := (int(ws.Col) - (len(msg) + 4))
			if !StripColor {
				fmt.Printf("%s %s%s\n", String(">>").Bold().Green(), msg, padding(buff))
			} else {
				fmt.Printf(">> %s%s\n", msg, padding(buff))
			}
		case msg := <-Error:
			buff := (int(ws.Col) - (len(msg) + 4))
			if !StripColor {
				fmt.Printf("%s %s%s\n", String("!!").Bold().Red(), msg, padding(buff))
			} else {
				fmt.Printf("!! %s%s\n", msg, padding(buff))
			}
		case msg := <-Debug:
			buff := (int(ws.Col) - (len(msg) + 4))
			if ShowDebug {
				if !StripColor {
					fmt.Printf("%s %s%s|\n", String("##").Bold().Cyan(), msg, padding(buff))
				} else {
					fmt.Printf("## %s%s\n", msg, padding(buff))
				}
			}
		case msg := <-Progress:
			buff := (int(ws.Col) - (len(msg) + 4))
			fmt.Printf("%s %s%s\r", String(">").Bold().Blue(), msg, padding(buff))
		case <-OutputSig:
			fmt.Println("signal")
		}
	}
}

func init() {
	Info = make(chan string)
	Error = make(chan string)
	Debug = make(chan string)
	Progress = make(chan string)
	OutputSig = make(chan os.Signal, 1)
}
