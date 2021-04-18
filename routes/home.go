package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/agext/levenshtein"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"

	"git.sr.ht/~humaid/linux-gen/builder"
	"git.sr.ht/~humaid/linux-gen/config"
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
	Script                 string
	Stringency             int
}

var useCasePkgs = make(map[string][]string)

var meta builder.SystemMetadata

var pkgMap = make(map[string]builder.SystemPackage)

var cmdlinePkgs = []string{"tmux", "imagemagick", "curl", "units", "inxi",
	"jq", "mlocate", "pv", "ranger", "sxiv", "feh", "screen", "nmap", "lm-sensors",
	"ffmpeg", "zathura", "zathura-pdf-poppler"}

type sectionBundle struct {
	SectionName string
	FAIcon      string
	Sections    []string
}

type kbdInfo struct {
	Code, Name string
}

var kbds = []kbdInfo{
	{Code: "us", Name: "USA"},
	{Code: "ad", Name: "Andorra"},
	{Code: "af", Name: "Afghanistan"},
	{Code: "ara", Name: "Arabic"},
	{Code: "al", Name: "Albania"},
	{Code: "am", Name: "Armenia"},
	{Code: "az", Name: "Azerbaijan"},
	{Code: "by", Name: "Belarus"},
	{Code: "be", Name: "Belgium"},
	{Code: "bd", Name: "Bangladesh"},
	{Code: "in", Name: "India"},
	{Code: "ba", Name: "Bosnia and Herzegovina"},
	{Code: "br", Name: "Brazil"},
	{Code: "bg", Name: "Bulgaria"},
	{Code: "ma", Name: "Morocco"},
	{Code: "mm", Name: "Myanmar"},
	{Code: "ca", Name: "Canada"},
	{Code: "cd", Name: "Congo, Democratic Republic of the"},
	{Code: "cn", Name: "China"},
	{Code: "hr", Name: "Croatia"},
	{Code: "cz", Name: "Czechia"},
	{Code: "dk", Name: "Denmark"},
	{Code: "nl", Name: "Netherlands"},
	{Code: "bt", Name: "Bhutan"},
	{Code: "ee", Name: "Estonia"},
	{Code: "ir", Name: "Iran"},
	{Code: "iq", Name: "Iraq"},
	{Code: "fo", Name: "Faroe Islands"},
	{Code: "fi", Name: "Finland"},
	{Code: "fr", Name: "France"},
	{Code: "gh", Name: "Ghana"},
	{Code: "gn", Name: "Guinea"},
	{Code: "ge", Name: "Georgia"},
	{Code: "de", Name: "Germany"},
	{Code: "gr", Name: "Greece"},
	{Code: "hu", Name: "Hungary"},
	{Code: "is", Name: "Iceland"},
	{Code: "il", Name: "Israel"},
	{Code: "it", Name: "Italy"},
	{Code: "jp", Name: "Japan"},
	{Code: "kg", Name: "Kyrgyzstan"},
	{Code: "kh", Name: "Cambodia"},
	{Code: "kz", Name: "Kazakhstan"},
	{Code: "la", Name: "Laos"},
	{Code: "latam", Name: "Latin American"},
	{Code: "lt", Name: "Lithuania"},
	{Code: "lv", Name: "Latvia"},
	{Code: "mao", Name: "Maori"},
	{Code: "me", Name: "Montenegro"},
	{Code: "mk", Name: "Macedonia"},
	{Code: "mt", Name: "Malta"},
	{Code: "mn", Name: "Mongolia"},
	{Code: "no", Name: "Norway"},
	{Code: "pl", Name: "Poland"},
	{Code: "pt", Name: "Portugal"},
	{Code: "ro", Name: "Romania"},
	{Code: "ru", Name: "Russia"},
	{Code: "rs", Name: "Serbia"},
	{Code: "si", Name: "Slovenia"},
	{Code: "sk", Name: "Slovakia"},
	{Code: "es", Name: "Spain"},
	{Code: "se", Name: "Sweden"},
	{Code: "ch", Name: "Switzerland"},
	{Code: "sy", Name: "Syria"},
	{Code: "tj", Name: "Tajikistan"},
	{Code: "lk", Name: "Sri Lanka"},
	{Code: "th", Name: "Thailand"},
	{Code: "tr", Name: "Turkey"},
	{Code: "tw", Name: "Taiwan"},
	{Code: "ua", Name: "Ukraine"},
	{Code: "gb", Name: "United Kingdom"},
	{Code: "uz", Name: "Uzbekistan"},
	{Code: "vn", Name: "Vietnam"},
	{Code: "kr", Name: "Korea, Republic of"},
	{Code: "jp", Name: "Japan (PC-98xx Series)"},
	{Code: "ie", Name: "Ireland"},
	{Code: "pk", Name: "Pakistan"},
	{Code: "mv", Name: "Maldives"},
	{Code: "za", Name: "South Africa"},
	{Code: "epo", Name: "Esperanto"},
	{Code: "np", Name: "Nepal"},
	{Code: "ng", Name: "Nigeria"},
	{Code: "et", Name: "Ethiopia"},
	{Code: "sn", Name: "Senegal"},
	{Code: "brai", Name: "Braille"},
	{Code: "tm", Name: "Turkmenistan"},
	{Code: "ml", Name: "Mali"},
	{Code: "tz", Name: "Tanzania"},
}

var sections = []sectionBundle{
	{
		SectionName: "Internet",
		FAIcon:      "fab fa-firefox-browser",
		Sections:    []string{"mail", "web"},
	},
	{
		SectionName: "Games",
		FAIcon:      "fas fa-puzzle-piece",
		Sections:    []string{"games"},
	},
	{
		SectionName: "Development",
		FAIcon:      "fas fa-code",
		Sections: []string{"python", "doc", "devel", "haskell", "database",
			"php", "java", "vcs", "javascript", "shells", "ocaml", "ruby",
			"lisp", "perl", "rust"},
	},
	{
		SectionName: "Education",
		FAIcon:      "fas fa-atom",
		Sections:    []string{"science", "math", "education"},
	},
	{
		SectionName: "Multimedia",
		FAIcon:      "fas fa-photo-video",
		Sections:    []string{"sound", "graphics", "video"},
	},
	{
		SectionName: "Miscellaneous",
		FAIcon:      "fas fa-boxes",
		Sections:    []string{"misc", "hamradio"},
	},
	{
		SectionName: "System Applications",
		FAIcon:      "fas fa-cog",
		Sections:    []string{"x11", "gnome", "kde", "xfce"},
	},
	{
		SectionName: "System Tools",
		FAIcon:      "fas fa-tools",
		Sections:    []string{"admin", "kernel", "net", "utils", "otherosfs"},
	},
	{
		SectionName: "System Libraries",
		FAIcon:      "fas fa-book",
		Sections:    []string{"libdevel", "libs", "oldlibs", "devel"},
	},
}

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

	// Get timezones
	now := time.Now()
	if tzFile, err := ioutil.ReadFile("tz"); err == nil {
		for _, line := range strings.Split(string(tzFile), "\n") {
			if loc, err := time.LoadLocation(line); err == nil && len(line) != 0 {
				then := now.In(loc)
				short, off := then.Zone()
				tzs = append(tzs, tzInfo{
					DBName:    line,
					ShortName: short,
					Offset:    off,
				})
			}
		}
	}
	sort.Sort(ByZone(tzs))
}

var tzs []tzInfo

type tzInfo struct {
	DBName    string
	ShortName string
	Offset    int
}

type ByZone []tzInfo

func (a ByZone) Len() int      { return len(a) }
func (a ByZone) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByZone) Less(i, j int) bool {
	return a[i].Offset < a[j].Offset
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

var filter = regexp.MustCompile(`^(multiverse/|universe/|restricted/)`)

func filterPackages(section sectionBundle) (pkgs []builder.SystemPackage) {
	for i, p := range meta.Packages {
		sect := filter.ReplaceAllString(p.Section, "")
		for _, bun := range section.Sections {
			if sect == bun {
				pkgs = append(pkgs, meta.Packages[i])
				break
			}
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

	if ctx.Req.Method == "POST" {

		var addPkgs, rmPkgs []string
		for i, p := range sessData.Packages {
			if p.Action == InstallPackage {
				addPkgs = append(addPkgs, sessData.Packages[i].Name)
			} else {
				rmPkgs = append(rmPkgs, sessData.Packages[i].Name)
			}
		}

		cust := builder.Customisation{
			AuthorID:        "web",
			DistName:        sessData.Name,
			DistVer:         sessData.Version,
			AddPackages:     addPkgs,
			RemovePackages:  rmPkgs,
			TZ:              sessData.Tz,
			Kbd:             sessData.Kbd,
			StringencyLevel: sessData.Stringency,
			Script:          sessData.Script,
		}

		c, err := json.Marshal(cust)

		req, err := http.NewRequest("POST", "http://localhost:8484/api", bytes.NewBuffer(c))
		if err != nil {
			config.Logger.Error(err)
			return
		}
		req.Header.Set("key", config.Config.SecretKey)

		/*AuthorID          string
		DistName, DistVer string
		AddPackages       []string
		RemovePackages    []string
		TZ, Kbd           string*/

		client := &http.Client{}
		client.Do(req)
		return
	}

	if stat, err := os.Stat("./builds/web/final.iso"); err == nil && !stat.IsDir() {
		ctx.Data["HasBuild"] = true
		ctx.Data["BuildTime"] = stat.ModTime()
	}

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

	if len(ctx.QueryTrim("rm")) > 0 {
		rm := ctx.QueryTrim("rm")
		for i, p := range sessData.Packages {
			if p.Name == rm {
				sessData.Packages = append(sessData.Packages[:i], sessData.Packages[i+1:]...)
				break
			}
		}

		sess.Set("sess", sessData)
		ctx.Redirect("/pkgs")
		return
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

	if len(ctx.QueryTrim("do")) > 0 {
		pkg := ctx.QueryTrim("do")

		// Check if exists in list
		for _, p := range sessData.Packages {
			if p.Name == pkg {
				ctx.Redirect("/pkgs")
				return
			}
		}

		for _, p := range meta.Packages {
			if p.Name == pkg {
				// Install if not installed, purge otherwise
				var action packageAction
				if p.Installed {
					action = PurgePackage
				} else {
					action = InstallPackage
				}
				sessData.Packages = append(sessData.Packages, sysPackage{
					Name:    p.Name,
					Version: p.Version,
					Action:  action,
				})

				break
			}
		}

		sess.Set("sess", sessData)

		ctx.Redirect("/pkgs")
		return
	}

	sec := ctx.QueryTrim("sec")
	if len(sec) > 0 {
		if sec == "all" {
			ctx.Data["Packages"] = meta.Packages
			ctx.Data["Cat"] = sectionBundle{
				SectionName: "View All Packages",
				FAIcon:      "fas fa-list-ul",
			}
		} else {
			for _, s := range sections {
				if sec == s.SectionName {
					pkgs := filterPackages(s)
					ctx.Data["Packages"] = pkgs
					ctx.Data["Cat"] = s
					break
				}
			}
		}
	} else if len(ctx.QueryTrim("q")) > 0 {
		query := ctx.QueryTrim("q")
		var pkgs []builder.SystemPackage
		for i, p := range meta.Packages {
			params := levenshtein.NewParams()
			params.BonusPrefix(16)
			params.BonusScale(0.9)
			if levenshtein.Similarity(p.Name, query, params) >= 0.5 {
				pkgs = append(pkgs, meta.Packages[i])
			}
		}
		ctx.Data["Packages"] = pkgs
		ctx.Data["Cat"] = sectionBundle{
			SectionName: "Search Results for \"" + query + "\"",
			FAIcon:      "fas fa-search",
		}
	}

	ctx.Data["Categories"] = sections

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
		sessData.Script = ctx.QueryTrim("script")
		sessData.Stringency = ctx.QueryInt("stringency")

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

	}

	ctx.Data["sess"] = sessData
	ctx.Data["kbds"] = kbds
	ctx.Data["tzs"] = tzs
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
