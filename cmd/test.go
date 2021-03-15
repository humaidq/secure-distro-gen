package cmd

import (
	"fmt"

	"git.sr.ht/~humaid/linux-gen/builder"
	"git.sr.ht/~humaid/linux-gen/config"

	"github.com/urfave/cli/v2"
)

// CmdTest represents the command that tests the builder.
var CmdTest = &cli.Command{
	Name:   "test",
	Usage:  "Run the build test",
	Action: test,
}

func test(clx *cli.Context) (err error) {
	config.LoadConfig()
	config.SetupLogger()
	defer config.Logger.Sync()

	cust := builder.Customisation{
		Author:      "Humaid",
		DistName:    "sifrOS",
		DistVer:     "0.1.0-test",
		AddPackages: []string{"sxiv"},
		TZ:          "Asia/Dubai",
		Kbd:         "us",
	}

	file, err := builder.Start(cust)
	if err != nil {
		panic(err)
	}
	fmt.Println("ISO:", file)

	return nil
}
