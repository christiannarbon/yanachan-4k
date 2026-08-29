package stats

import (
	"testing"
	"time"

	"github.com/christiannarbon/yanachan-4k/backend/internal/github"
)

// The window is drawn in the server's own zone, so the tests fix one that is
// not UTC. A pull request merged at 23:50 in Tokyo belongs to that Tokyo day,
// not to the UTC day it happens to fall in.
var tokyo = time.FixedZone("Asia/Tokyo", 9*3600)

func at(y int, m time.Month, d, h int) time.Time {
	return time.Date(y, m, d, h, 0, 0, 0, tokyo)
}

func TestResolveWeekCoversWholeLocalDays(t *testing.T) {
	w := ResolveWeek(at(2026, time.August, 30, 14), 7)

	if w.Days != 7 {
		t.Errorf("days = %d, want 7", w.Days)
	}
	if got, want := w.Since.Format(time.RFC3339), "2026-08-24T00:00:00+09:00"; got != want {
		t.Errorf("since = %s, want %s", got, want)
	}
	if h, m, s := w.Since.Clock(); h != 0 || m != 0 || s != 0 {
		t.Errorf("since is not midnight: %v", w.Since)
	}
}

func TestResolveWeekClampsDays(t *testing.T) {
	now := at(2026, time.August, 30, 14)
	if got := ResolveWeek(now, 0).Days; got != DefaultDays {
		t.Errorf("zero days = %d, want the default %d", got, DefaultDays)
	}
	if got := ResolveWeek(now, -3).Days; got != DefaultDays {
		t.Errorf("negative days = %d, want the default %d", got, DefaultDays)
	}
	if got := ResolveWeek(now, 5_000).Days; got != MaxDays {
		t.Errorf("oversized days = %d, want the cap %d", got, MaxDays)
	}
}

func TestEmptyDaysIsOneColumnPerDay(t *testing.T) {
	days := emptyDays(ResolveWeek(at(2026, time.August, 30, 14), 7))

	if len(days) != 7 {
		t.Fatalf("got %d columns, want 7", len(days))
	}
	if days[0].Date != "2026-08-24" {
		t.Errorf("first column = %s, want 2026-08-24", days[0].Date)
	}
	if days[6].Date != "2026-08-30" {
		t.Errorf("last column = %s, want 2026-08-30", days[6].Date)
	}
	for _, d := range days {
		if d.Total() != 0 {
			t.Errorf("%s starts at %d, want an empty column", d.Date, d.Total())
		}
	}
}

func TestDayKeyBucketsInTheWindowsZone(t *testing.T) {
	w := ResolveWeek(at(2026, time.August, 30, 14), 7)
	// 23:50 in Tokyo is 14:50 the same day in UTC; both must land on the 29th.
	late := at(2026, time.August, 29, 23).Add(50 * time.Minute)

	if got := dayKey(late, w); got != "2026-08-29" {
		t.Errorf("local stamp bucketed to %s, want 2026-08-29", got)
	}
	if got := dayKey(late.UTC(), w); got != "2026-08-29" {
		t.Errorf("same instant in UTC bucketed to %s, want 2026-08-29", got)
	}
}

func TestRhythmCountsActiveDaysAndTheRun(t *testing.T) {
	cases := []struct {
		name       string
		totals     []int
		wantActive int
		wantStreak int
	}{
		{"a quiet week", []int{0, 0, 0, 0, 0, 0, 0}, 0, 0},
		{"every day", []int{1, 2, 1, 3, 1, 1, 2}, 7, 7},
		// The streak is allowed to survive an empty today: the dashboard is
		// opened in the morning, before anything has been merged.
		{"today has not started", []int{0, 1, 1, 1, 1, 1, 0}, 5, 5},
		{"broken midweek", []int{1, 1, 0, 0, 1, 1, 1}, 5, 3},
		{"yesterday empty too", []int{1, 1, 1, 1, 1, 0, 0}, 5, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			days := make([]Day, len(tc.totals))
			for i, n := range tc.totals {
				days[i] = Day{Opened: n}
			}
			active, streak := rhythm(days)
			if active != tc.wantActive {
				t.Errorf("active = %d, want %d", active, tc.wantActive)
			}
			if streak != tc.wantStreak {
				t.Errorf("streak = %d, want %d", streak, tc.wantStreak)
			}
		})
	}
}

func TestTopRepoBreaksTiesByName(t *testing.T) {
	name, n := topRepo(map[string]int{"zed/one": 3, "acme/two": 3, "acme/one": 1})
	if name != "acme/two" || n != 3 {
		t.Errorf("top repo = %q (%d), want acme/two (3)", name, n)
	}
	if name, n := topRepo(map[string]int{}); name != "" || n != 0 {
		t.Errorf("empty week = %q (%d), want the zero value", name, n)
	}
}

func TestInWeekExcludesEitherSide(t *testing.T) {
	w := ResolveWeek(at(2026, time.August, 30, 14), 7)
	if inWeek(w.Since.Add(-time.Second), w) {
		t.Error("a stamp before the window counted as inside it")
	}
	if !inWeek(w.Since, w) {
		t.Error("the first instant of the window counted as outside it")
	}
	if inWeek(w.Until.Add(time.Hour), w) {
		t.Error("a stamp in the future counted as inside the window")
	}
}

// review builds one submitted review, which is what the reviewed search is
// reduced over.
func review(login, state string, when time.Time) github.ReviewNode {
	return github.ReviewNode{CreatedAt: when, State: state, Author: &github.Actor{Typename: "User", Login: login}}
}

func TestReviewCountingIgnoresOthersAndDrafts(t *testing.T) {
	w := ResolveWeek(at(2026, time.August, 30, 14), 7)
	pr := github.PullRequestStat{Number: 4, URL: "https://x/4"}
	pr.Reviews.Nodes = []github.ReviewNode{
		review("me", "APPROVED", at(2026, time.August, 28, 10)),
		review("me", "COMMENTED", at(2026, time.August, 28, 11)),
		// A second day on the same branch: two reviews, two active columns,
		// but still one pull request reviewed.
		review("me", "COMMENTED", at(2026, time.August, 29, 9)),
		// Not mine.
		review("alice", "APPROVED", at(2026, time.August, 28, 12)),
		// Still a draft in somebody's browser: never submitted.
		review("me", "PENDING", at(2026, time.August, 29, 10)),
		// Mine, but from before the window.
		review("me", "APPROVED", at(2026, time.August, 20, 9)),
	}

	written, approvals, days := countReviews(pr, "me", w)
	if written != 3 {
		t.Errorf("reviews written = %d, want 3", written)
	}
	if approvals != 1 {
		t.Errorf("approvals = %d, want 1", approvals)
	}
	if len(days) != 2 {
		t.Errorf("active days = %v, want two", days)
	}
}
