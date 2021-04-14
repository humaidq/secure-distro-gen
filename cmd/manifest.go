package cmd

import (
	"fmt"

	"git.sr.ht/~humaid/linux-gen/builder"
	"git.sr.ht/~humaid/linux-gen/config"

	"github.com/urfave/cli/v2"
)

// CmdMetadata represents the command that gets the metadata.
var CmdMetadata = &cli.Command{
	Name:   "metadata",
	Usage:  "Gets the metadata",
	Action: metadata,
}

func metadata(clx *cli.Context) (err error) {
	config.LoadConfig()
	config.SetupLogger()
	defer config.Logger.Sync()

	file, err := builder.GetMetadata()
	if err != nil {
		panic(err)
	}
	fmt.Println(file)

	return nil
}
