package wizard

// Question represents a single question in the wizard.
type Question struct {
	QuestionID  uint32
	Question    string
	Description string
	Options     []Option
}

// Option represents an option which a Question might have.
type Option struct {
	OptionID string // used in forms
	Option   string
}

// QuestionResponse represents a response picked by a user from the online
// wizard, it contains the question ID and the chosen option ID.
type QuestionResponse struct {
	QuestionID     uint32
	ChosenOptionID string
}

// QuestionMap is a map of Questions, where the key is the QuestionID.
var QuestionMap map[uint32]*Question

func init() {
	// Map the Questions array to QuestionMap
	QuestionMap = make(map[uint32]*Question)
	for i := range Questions {
		QuestionMap[Questions[i].QuestionID] = &Questions[i]
	}
}

// Questions is the array of available questions.
var Questions = []Question{
	{
		QuestionID:  1,
		Question:    "Who is this for?",
		Description: "This will help determine the type of questions",
		Options: []Option{
			{OptionID: "myself", Option: "For myself"},
			{OptionID: "someone", Option: "For someone else"},
			{OptionID: "org", Option: "For my organisation"},
			{OptionID: "other", Option: "Other"},
		},
	},
	{
		QuestionID:  2,
		Question:    "What is your use case?",
		Description: "This will determine which software will be bundled with the system by default",
		Options: []Option{
			{OptionID: "school", Option: "Student/School"},
			{OptionID: "office", Option: "Office"},
			{OptionID: "studio", Option: "Studio (Audio, Graphics, Video)"},
			{OptionID: "cs", Option: "Computer Scientist/Developer"},
			{OptionID: "other", Option: "Other"},
		},
	},
	{
		QuestionID:  3,
		Question:    "How do you rate your Linux knowledge?",
		Description: "This will determine the complexity of questions",
		Options: []Option{
			{OptionID: "beginner", Option: "Beginner: I need help with some/most tasks"},
			{OptionID: "moderate", Option: "Moderate: I know how to manage most things myself"},
			{OptionID: "expert", Option: "Expert: I usually know how to fix things myself"},
		},
	},
	{
		QuestionID:  4,
		Question:    "What is your security risk factor?",
		Description: "Depending on the severity, more security measures might be added",
		Options: []Option{
			{OptionID: "low", Option: "Low: I am not a targeted individual"},
			{OptionID: "medium", Option: "Medium: I am in a mediium risk environment"},
			{OptionID: "high", Option: "High: I am potentially/currently being targetted"},
		},
	},
	{
		QuestionID:  5,
		Question:    "How much do you care about usability",
		Description: "This comes into consideration when configuring security measures that may affect usability",
		Options: []Option{
			{OptionID: "high", Option: "High: I really care about usability, and I don't mind if it affects usability"},
			{OptionID: "medium", Option: "Medium: I don't really mind if some security measures affect usability"},
			{OptionID: "low", Option: "Low: I don't care if security drastically affects security"},
		},
	},
	{
		QuestionID:  6,
		Question:    "Do you prefer to use the shell/terminal?",
		Description: "This will determine whether more terminal-based utility tools will be bundled with the system",
		Options: []Option{
			{OptionID: "yes", Option: "Yes, I prefer to use the terminal"},
			{OptionID: "dontmind", Option: "I don't mind using the terminal"},
			{OptionID: "no", Option: "No, I'd always prefer to use a graphical user interface"},
		},
	},
	{
		QuestionID:  7,
		Question:    "Will you be running binaries or scripts downloaded directly from the web?",
		Description: "This will determine whether the home partition will have the no-execute flag",
		Options: []Option{
			{OptionID: "no", Option: "No, I'll only stick to the official system packages and repositories"},
			{OptionID: "yes", Option: "Yes, I'd like to be able to execute files downloaded from the web"},
		},
	},
}
