package cmd

import (
	"git.sr.ht/~humaid/linux-gen/builder"
	"git.sr.ht/~humaid/linux-gen/config"

	"github.com/urfave/cli/v2"
)

// CmdAPI represents a command-line command
// which starts the API server.
var CmdAPI = &cli.Command{
	Name:   "api",
	Usage:  "Start the API server",
	Action: api,
}

func api(clx *cli.Context) (err error) {
	config.LoadConfig()
	config.SetupLogger()
	defer config.Logger.Sync()

	builder.RunAPI()
	return nil
}
