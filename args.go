package main

import (
	"fmt"
	"os"
	"time"

	"git.thwap.org/splat/gout"
	"github.com/urfave/cli"
)

// Putting this in the global scope let's the other source
// files fill it in with their init() functions.

var (
	Commands   []cli.Command
	StartTime  int64
	StripColor bool
	ShowDebug  bool
)

// Starting things rolling down the hill
func main() {
	// Setup our output
	gout.Setup(false, false, true, "")
	gout.Output.Throbber = []string{
		string(gout.String(".").Cyan()),
		string(gout.String("o").Bold().Cyan()),
		string(gout.String("O").Bold().White()),
		string(gout.String("o").Bold().Cyan()),
	}
	gout.Output.Prompts["info"] = fmt.Sprintf("%s%s%s",
		gout.String(".").Cyan(),
		gout.String(".").Bold().Cyan(),
		gout.String(".").Bold().White())
	gout.Output.Prompts["warn"] = fmt.Sprintf("%s%s%s",
		gout.String(".").Yellow(),
		gout.String(".").Bold().Yellow(),
		gout.String(".").Bold().White())
	gout.Output.Prompts["error"] = fmt.Sprintf("%s%s%s",
		gout.String(".").Red(),
		gout.String(".").Bold().Red(),
		gout.String(".").Bold().White())
	gout.Output.Prompts["debug"] = fmt.Sprintf("%s%s%s",
		gout.String(".").Purple(),
		gout.String(".").Bold().Purple(),
		gout.String(".").Bold().White())
	gout.Output.Prompts["status"] = gout.Output.Prompts["info"]

	// Define our app container
	app := cli.NewApp()

	// And give all the metadata a nice sound setting.
	app.Name = "flustro"
	app.Usage = "A whisper file toolkit."
	app.Version = "0.1.0"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Mike 'Fuzzy' Partin",
			Email: "fuzzy@fumanchu.org",
		},
	}
	app.Copyright = "(c) 2016 Mike 'Fuzzy' Partin"
	app.Commands = Commands
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "d", Usage: "Show debugging messages", Destination: &ShowDebug},
		cli.BoolFlag{Name: "c", Usage: "Disable colors in output", Destination: &StripColor},
	}

	// Now let's do things
	StartTime = time.Now().Unix()
	app.Run(os.Args)
	time.Sleep(500 * time.Millisecond)
}
