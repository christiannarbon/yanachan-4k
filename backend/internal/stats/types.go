// Package stats turns a week of GitHub search results into the landing
// dashboard: what the viewer opened, merged, closed and reviewed, the shape of
// each day, and a few things worth being pleased about.
//
// Where package board answers "what needs me right now", this answers "what
// did I get done". The two share a client and a login and nothing else.
package stats

import "time"

// Kcal weights.
//
// The dashboard leads with a calorie figure, which is the joke this repository
// is named after -- 4K is a calorie count, and the theme paints the total in
// the calorie meter's yellow tiles. The weights are arbitrary by definition,
// so they are at least arbitrary in one place, and ordered the way the work
// is: shipping beats opening, reviewing somebody else's branch counts, and a
// pull request you closed unmerged still cost you the afternoon.
const (
	KcalPerOpened   = 200
	KcalPerMerged   = 400
	KcalPerClosed   = 100
	KcalPerReview   = 150
	KcalPerApproval = 50
)

// Week is the reporting window: whole local days, ending with today.
type Week struct {
	Days  int       `json:"days"`
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

// Day is one column of the activity chart. Reviewed counts distinct pull
// requests reviewed that day, not reviews submitted, so a day where you went
// back and forth on one branch reads as one.
type Day struct {
	Date     string `json:"date"` // YYYY-MM-DD in the server's zone
	Opened   int    `json:"opened"`
	Merged   int    `json:"merged"`
	Reviewed int    `json:"reviewed"`
}

// Total is the day's height in the chart.
func (d Day) Total() int { return d.Opened + d.Merged + d.Reviewed }

// PRRef is the little a highlight needs to link back to a pull request.
type PRRef struct {
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// Highlights are the week's superlatives. Every one of them is optional: a
// week with nothing merged simply has no fastest merge, and the frontend
// leaves the row out rather than printing a zero.
type Highlights struct {
	FastestMerge   *PRRef `json:"fastestMerge,omitempty"`
	FastestMinutes int    `json:"fastestMinutes"`
	BiggestMerge   *PRRef `json:"biggestMerge,omitempty"`
	BiggestLines   int    `json:"biggestLines"`
	TopRepo        string `json:"topRepo"`
	TopRepoCount   int    `json:"topRepoCount"`
}

// Stats is the whole payload handed to the dashboard tab.
type Stats struct {
	Login       string    `json:"login"`
	Week        Week      `json:"week"`
	GeneratedAt time.Time `json:"generatedAt"`

	// The four headline counts, in the order the tiles show them.
	Opened   int `json:"opened"`
	Merged   int `json:"merged"`
	Closed   int `json:"closed"`   // closed without merging
	Reviewed int `json:"reviewed"` // distinct pull requests you reviewed

	ReviewsWritten int `json:"reviewsWritten"`
	Approvals      int `json:"approvals"`
	Repos          int `json:"repos"`
	Additions      int `json:"additions"`
	Deletions      int `json:"deletions"`
	FilesChanged   int `json:"filesChanged"`

	Kcal       int `json:"kcal"`
	ActiveDays int `json:"activeDays"`
	Streak     int `json:"streak"`

	Daily      []Day      `json:"daily"`
	Highlights Highlights `json:"highlights"`
	Warning    string     `json:"warning,omitempty"`
}
