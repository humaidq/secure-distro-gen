package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"

	"git.sr.ht/~humaid/linux-gen/builder"
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

type packageAction int

const (
	InstallPackage = iota
	PurgePackage
)

type sysPackage struct {
	Name, Version string
	Action        packageAction
}

type sessionData struct {
	Name, Version, Kbd, Tz string
	Packages               []sysPackage
	Processed              bool
	Stringency             int
}

var useCasePkgs = make(map[string][]string)

var meta builder.SystemMetadata

var pkgMap = make(map[string]builder.SystemPackage)

var cmdlinePkgs = []string{"tmux", "imagemagick", "curl", "units", "inxi",
	"jq", "mlocate", "pv", "ranger", "sxiv", "feh", "screen", "nmap", "lm-sensors",
	"ffmpeg", "zathura", "zathura-pdf-poppler"}

func init() {
	if data, err := ioutil.ReadFile("./metadata.json"); err == nil {
		if err := json.Unmarshal(data, &meta); err == nil {
			for i, p := range meta.Packages {
				pkgMap[p.Name] = meta.Packages[i]
			}
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}

	useCasePkgs["school"] = []string{"gcompris", "kdeedu", "geogebra",
		"scratch", "tuxmath", "gnome-dictionary", "inkscape", "mypaint", "gimp",
		"xournal", "gbrainy", "tuxtype", "epoptes", "epoptes-client", "calibre",
		"vym", "freeplane", "gnome-sound-recorder", "audacity", "rocs", "atomix",
		"anki"}

	useCasePkgs["office"] = []string{"gnucash", "ofxstatement",
		"ofxstatement-plugins", "gimp", "inkscape", "gnome-dictionary",
		"xournal", "audacity", "thunderbird", "timeshift"}

	useCasePkgs["studio"] = []string{"obs-studio", "audacity",
		"kdenlive", "inkscape", "mypaint", "handbrake", "lmms",
		"blender", "youtube-dl"}

	useCasePkgs["cs"] = []string{"clang", "make", "cmake",
		"bat", "audacity", "gimp", "inkscape", "mypaint", "ffmpeg",
		"bvi", "curl", "dust", "feh", "ffmpeg", "build-essential",
		"licensor", "jq", "pandoc", "plantuml", "shellcheck", "youtube-dl"}

	useCasePkgs["other"] = []string{}
}

func getPackages(list []string) (s []sysPackage) {
	for _, p := range list {
		fp, ok := pkgMap[p]
		if ok {
			s = append(s, sysPackage{
				Name:    fp.Name,
				Version: fp.Version,
				Action:  InstallPackage,
			})
		}
	}

	return
}

// BuildHandler response for the final page.
func BuildHandler(ctx *macaron.Context, sess session.Store) {
	if !hasDoneWizard(sess) {
		ctx.Redirect("/")
		return
	}

	var sessData sessionData
	if s := sess.Get("sess"); s != nil {
		sessData = s.(sessionData)
	}
	ctx.Data["sess"] = sessData

	ctx.HTML(200, "build")
}

// PackagesHandler response for the final page.
func PackagesHandler(ctx *macaron.Context, sess session.Store) {
	if !hasDoneWizard(sess) {
		ctx.Redirect("/")
		return
	}

	var sessData sessionData
	if s := sess.Get("sess"); s != nil {
		sessData = s.(sessionData)
	}
	ctx.Data["sess"] = sessData

	ctx.HTML(200, "packages")
}

func getCategories() (categories []string) {
	has := make(map[string]bool)

	for _, p := range meta.Packages {
		if !has[p.Section] {
			categories = append(categories, p.Section)
			has[p.Section] = true
		}
	}

	return
}

// SelectPackageHandler response for the final page.
func SelectPackageHandler(ctx *macaron.Context, sess session.Store) {
	if !hasDoneWizard(sess) {
		ctx.Redirect("/")
		return
	}

	var sessData sessionData
	if s := sess.Get("sess"); s != nil {
		sessData = s.(sessionData)
	}
	ctx.Data["sess"] = sessData
	ctx.Data["meta"] = meta

	ctx.Data["Categories"] = getCategories()

	ctx.HTML(200, "selectpkg")
}

// ConfigHandler response for the final page.
func ConfigHandler(ctx *macaron.Context, sess session.Store) {
	if !hasDoneWizard(sess) {
		ctx.Redirect("/")
		return
	}

	// TODO Make sure this exists somewhere else
	var sessData sessionData
	if s := sess.Get("sess"); s != nil {
		sessData = s.(sessionData)
	}

	if ctx.Req.Method == "POST" {
		// TODO check if not empty
		sessData.Name = ctx.QueryTrim("name")
		sessData.Version = ctx.QueryTrim("ver")
		sessData.Kbd = ctx.QueryTrim("kbd")
		sessData.Tz = ctx.QueryTrim("tz")

		sess.Set("sess", sessData)
		ctx.Redirect("/config")
		return
	}

	if !sessData.Processed {
		answers := sess.Get("answers").(*[]wizard.QuestionResponse)
		score := 0
		for _, a := range *answers {
			switch a.QuestionID {
			case 2:
				sessData.Packages = append(sessData.Packages,
					getPackages(useCasePkgs[a.ChosenOptionID])...)
			case 4:
				switch a.ChosenOptionID {
				case "medium":
					score += 1
				case "high":
					score += 2
				}
			case 5:
				switch a.ChosenOptionID {
				case "medium":
					score += 1
				case "low":
					score += 2
				}
			case 6:
				if a.ChosenOptionID == "yes" {
					sessData.Packages = append(sessData.Packages,
						getPackages(cmdlinePkgs)...)
				}
			}

			sessData.Stringency = score
		}
		sessData.Processed = true

		sess.Set("sess", sessData)
		ctx.Data["sess"] = sessData

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

func hasDoneWizard(sess session.Store) bool {
	answers := sess.Get("answers")
	if answers == nil {
		return false
	}

	a := answers.(*[]wizard.QuestionResponse)
	q := getLastOption(*a) + 1
	if uint32(len(wizard.Questions)) != q {
		return false
	}

	return true
}

// WizardHandler response for the wizard question page.
func WizardHandler(ctx *macaron.Context, sess session.Store) {
	if sess.Get("answers") == nil {
		resps := new([]wizard.QuestionResponse)
		sess.Set("answers", resps)
	}
	answers := sess.Get("answers").(*[]wizard.QuestionResponse)
	q := getLastOption(*answers) + 1

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
		exists := false
		for _, o := range wizard.QuestionMap[q].Options {
			if o.OptionID == ctx.QueryTrim("option") {
				exists = true
				break
			}
		}

		if !exists {
			ctx.Redirect("/wizard")
			return
		}

		opt := wizard.QuestionResponse{
			QuestionID:     qidForm,
			ChosenOptionID: ctx.QueryTrim("option"),
		}
		newAns := append(*answers, opt)
		sess.Set("answers", &newAns)
		sess.Get("answers")
		ctx.Redirect("/wizard")
	}

	if uint32(len(wizard.Questions)) == q {

		ctx.Redirect("/config")
	}

	ctx.Data["Question"] = wizard.QuestionMap[q]
	ctx.HTML(200, "wizard")
}
