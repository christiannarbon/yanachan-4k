package board

import "time"

// Status classifies an entry's activity inside the window.
const (
	StatusReply = "reply" // somebody answered after your last comment
	StatusNew   = "new"   // new activity, you had not commented yet
	StatusQuiet = "quiet" // nothing new in the window
)

// Section kinds drive which tab an entry lands in.
const (
	KindMine   = "mine"
	KindReview = "review"
	KindTeam   = "team"
	KindOrg    = "org"
)

// Entry is one pull request as rendered on the board.
type Entry struct {
	Number         int        `json:"number"`
	Title          string     `json:"title"`
	URL            string     `json:"url"`
	Repo           string     `json:"repo"`
	IsDraft        bool       `json:"isDraft"`
	Author         string     `json:"author"`
	AuthorIsBot    bool       `json:"authorIsBot"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	Checks         string     `json:"checks"`         // success | failure | pending | ""
	ReviewDecision string     `json:"reviewDecision"` // approved | changes_requested | review_required | ""
	Active         bool       `json:"active"`
	Hot            bool       `json:"hot"`
	Status         string     `json:"status"`
	HumanActors    []string   `json:"humanActors"`
	BotActors      []string   `json:"botActors"`
	HumanCount     int        `json:"humanCount"`
	BotCount       int        `json:"botCount"`
	LastActivityAt *time.Time `json:"lastActivityAt"`
	// Review-queue fields.
	Touched              bool       `json:"touched"`
	YourState            string     `json:"yourState"` // approved | changes_requested | commented | ""
	YourLastAt           *time.Time `json:"yourLastAt"`
	Awaiting             string     `json:"awaiting"` // you | team | ""
	AlsoRequestedFromYou bool       `json:"alsoRequestedFromYou"`
}

// Section is one tab on the dashboard.
type Section struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Kind    string  `json:"kind"`
	Ref     string  `json:"ref"` // team slug or org login, empty for the built-in sections
	Entries []Entry `json:"entries"`
	Total   int     `json:"total"`
	Active  int     `json:"active"`
	Hot     int     `json:"hot"`
	Error   string  `json:"error,omitempty"`
}

// Board is the whole payload handed to the frontend.
type Board struct {
	Login       string    `json:"login"`
	AuthMode    string    `json:"authMode"`
	Window      Window    `json:"window"`
	Sections    []Section `json:"sections"`
	GeneratedAt time.Time `json:"generatedAt"`
	OnlyActive  bool      `json:"onlyActive"`
	Limit       int       `json:"limit"`
	Warning     string    `json:"warning,omitempty"`
}
