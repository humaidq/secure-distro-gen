package cmd

import (
	"fmt"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"

	"git.sr.ht/~humaid/linux-gen/config"
	"git.sr.ht/~humaid/linux-gen/routes"

	"html/template"
	"log"
	"net/http"

	_ "github.com/go-macaron/session/mysql" // MySQL driver for persistent sessions
	"github.com/urfave/cli/v2"
)

// CmdStart represents a command-line command
// which starts the bot.
var CmdStart = &cli.Command{
	Name:    "run",
	Aliases: []string{"start", "web"},
	Usage:   "Start the web server",
	Action:  start,
}

func start(clx *cli.Context) (err error) {
	config.LoadConfig()
	//engine := models.SetupEngine()
	//defer engine.Close()

	// Run macaron
	m := macaron.Classic()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Funcs: []template.FuncMap{map[string]interface{}{
			"Len": func(arr []string) int {
				return len(arr)
			},
		}},
		IndentJSON: true,
	}))

	sessOpt := session.Options{
		CookieLifeTime: 2629744, // 1 month
		Gclifetime:     3600,    // gc every 1 hour
		CookieName:     "hithereimacookie",
	}

	m.Use(session.Sessioner(sessOpt))
	m.Use(csrf.Csrfer())
	m.Use(routes.ContextInit())

	m.Get("/", routes.HomepageHandler)
	m.Get("/wizard", routes.WizardHandler)
	m.Post("/wizard", routes.WizardHandler)
	m.Get("/final", routes.FinalHandler)
	m.Post("/final", routes.FinalHandler)
	m.Get("/download", routes.DownloadHandler)

	log.Printf("Starting web server on port %s\n", config.Config.SitePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Config.SitePort), m))
	return nil
}
