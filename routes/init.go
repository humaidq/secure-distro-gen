package routes

import (
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"

	"git.sr.ht/~humaid/linux-gen/config"
)

// ContextInit is a middleware which initialises some global variables.
func ContextInit() macaron.Handler {
	return func(ctx *macaron.Context, sess session.Store, f *session.Flash) {
		ctx.Data["SiteTitle"] = config.Config.SiteName
	}
}
