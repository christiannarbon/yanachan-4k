package board

import (
	"fmt"
	"time"
)

// Window kinds. The frontend translates the window from Kind and Hours; Label
// is the same sentence in English, kept for clients that do not.
const (
	WindowFixed       = "fixed"
	WindowBusinessDay = "business-day"
)

// Window is the activity window: everything newer than Cutoff counts as new.
type Window struct {
	// Kind is machine-readable: WindowFixed or WindowBusinessDay.
	Kind   string    `json:"kind"`
	Label  string    `json:"label"`
	Hours  int       `json:"hours"`
	Cutoff time.Time `json:"cutoff"`
	Now    time.Time `json:"now"`
}

// ResolveWindow reproduces the shell script's "last business day" rule.
//
//	Tue-Fri -> previous 24h
//	Sat     -> back to Fri (24h)
//	Sun     -> back to Fri (48h)
//	Mon     -> back to Fri (72h)
//
// A positive hoursOverride replaces the rule with a fixed window.
func ResolveWindow(now time.Time, hoursOverride int) Window {
	if hoursOverride > 0 {
		return Window{
			Kind:   WindowFixed,
			Label:  fmt.Sprintf("last %dh", hoursOverride),
			Hours:  hoursOverride,
			Cutoff: now.Add(-time.Duration(hoursOverride) * time.Hour),
			Now:    now,
		}
	}
	days := 1
	kind, label := WindowFixed, "last 24h"
	switch now.Weekday() {
	case time.Monday:
		days, kind, label = 3, WindowBusinessDay, "since last business day (Fri)"
	case time.Saturday:
		days, kind, label = 1, WindowBusinessDay, "since last business day (Fri)"
	case time.Sunday:
		days, kind, label = 2, WindowBusinessDay, "since last business day (Fri)"
	}
	return Window{
		Kind:   kind,
		Label:  label,
		Hours:  days * 24,
		Cutoff: now.AddDate(0, 0, -days),
		Now:    now,
	}
}
