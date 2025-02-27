package main

import (
	"log"
	"os"

	"git.sr.ht/~humaid/linux-gen/cmd"

	"github.com/urfave/cli/v2"
)

// VERSION specifies the version
var VERSION = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "linux-gen"
	app.Usage = "Custom Secure Linux Distribution Generator"
	app.Version = VERSION
	app.Commands = []*cli.Command{
		cmd.CmdStart,
		cmd.CmdTest,
		cmd.CmdAPI,
		cmd.CmdMetadata,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
