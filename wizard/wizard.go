package wizard

type Question struct {
	QuestionID uint32
	Question   string
	Options    []Option
}

type Option struct {
	OptionID string // used in forms
	Option   string
}

var questions = []Question{
	{
		QuestionID: 1,
		Question:   "Who is this for?",
		Options: []Option{
			{OptionID: "myself", Option: "For myself"},
			{OptionID: "someone", Option: "For someone else"},
			{OptionID: "org", Option: "For my organisation"},
			{OptionID: "other", Option: "Other"},
		},
	},
	{
		QuestionID: 2,
		Question:   "What is your use case?",
		Options: []Option{
			{OptionID: "school", Option: "Student/School"},
			{OptionID: "office", Option: "Office"},
			{OptionID: "studio", Option: "Studio (Audio, Graphics, Video)"},
			{OptionID: "cs", Option: "Computer Scientist/Developer"},
			{OptionID: "other", Option: "Other"},
		},
	},
}
