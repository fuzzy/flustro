package main

import (
	"os"
	"time"

	"github.com/urfave/cli"
)

// Putting this in the global scope let's the other source
// files fill it in with their init() functions.

var Commands []cli.Command

// Starting things rolling down the hill
func main() {
	// Define our app container
	app := cli.NewApp()

	// And give all the metadata a nice sound setting.
	app.Name = "flustro"
	app.Usage = "A whisper file toolkit."
	app.Version = "0.1.0"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Mike 'Fuzzy' Partin",
			Email: "fuzzy@fumanchu.org",
		},
	}
	app.Copyright = "(c) 2016 Mike 'Fuzzy' Partin"
	app.Commands = Commands

	// Now let's do things
	app.Run(os.Args)
}