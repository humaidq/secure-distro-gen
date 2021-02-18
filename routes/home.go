package routes

import (
	"fmt"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"

	"git.sr.ht/~humaid/linux-gen/wizard"
)

// HomepageHandler response for the home page.
func HomepageHandler(ctx *macaron.Context, sess session.Store) {
	ctx.Data["IsHome"] = 1
	ctx.HTML(200, "index")
}

// DownloadHandler response for the download page.
func DownloadHandler(ctx *macaron.Context, sess session.Store) {
	ctx.HTML(200, "download")
}

// FinalHandler response for the final page.
func FinalHandler(ctx *macaron.Context, sess session.Store) {
	if ctx.Req.Method == "POST" {
		sess.Set("os-name", ctx.QueryTrim("name"))
		sess.Set("os-ver", ctx.QueryTrim("ver"))
		ctx.Redirect("/download")
		return
	}
	ctx.HTML(200, "name")
}

func getLastOption(resps []wizard.QuestionResponse) (last uint32) {
	for _, r := range resps {
		if r.QuestionID > last {
			last = r.QuestionID
		}
	}
	return
}

// WizardHandler response for the wizard question page.
func WizardHandler(ctx *macaron.Context, sess session.Store) {
	if sess.Get("answers") == nil {
		resps := new([]wizard.QuestionResponse)
		sess.Set("answers", resps)
	}
	answers := sess.Get("answers").(*[]wizard.QuestionResponse)
	q := getLastOption(*answers) + 1

	fmt.Println("len", len(*answers))

	if ctx.Req.Method == "POST" {
		qidForm := uint32(ctx.QueryInt("qid"))

		if q != qidForm {
			fmt.Println("not correct ordering, reset!")
			sess.Set("answers", new([]wizard.QuestionResponse))
			ctx.Redirect("/?err=0")
			return
		}
		//qdata := wizard.QuestionMap[qidForm]

		// TODO verify question ID sequence
		// TODO verify option ID

		opt := wizard.QuestionResponse{
			QuestionID:     qidForm,
			ChosenOptionID: ctx.Query("option"),
		}
		newAns := append(*answers, opt)
		fmt.Println("setting to ", newAns)
		fmt.Println(sess.Set("answers", &newAns))
		fmt.Println(sess.Get("answers"))
		ctx.Redirect("/wizard")
	}

	if uint32(len(wizard.Questions)) == q {
		ctx.Redirect("/final")
	}
	fmt.Println("Displaying option", q)

	ctx.Data["Question"] = wizard.QuestionMap[q]
	ctx.HTML(200, "wizard")
}
