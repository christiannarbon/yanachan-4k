package stats

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/christiannarbon/yanachan-4k/backend/internal/github"
)

// DefaultDays is "the last week": today plus the six days before it.
const DefaultDays = 7

// MaxDays bounds what the API will accept, so one query cannot ask GitHub for
// a year of history inside a request the browser is waiting on.
const MaxDays = 90

type Request struct {
	Login string
	Days  int
	Now   time.Time
	Limit int
}

// ResolveWeek returns whole local days ending with the one now falls in.
//
// Whole days, not a rolling 168 hours: the dashboard draws a column per day
// and labels it with a weekday, so a window that started at 14:07 last Sunday
// would put two half-Sundays at the ends of the chart and make the first and
// last columns lie about themselves.
func ResolveWeek(now time.Time, days int) Week {
	if days < 1 {
		days = DefaultDays
	}
	if days > MaxDays {
		days = MaxDays
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return Week{
		Days:  days,
		Since: start.AddDate(0, 0, -(days - 1)),
		Until: now,
	}
}

const searchBase = "is:pr"

// Build runs the week's four searches in one round trip and reduces them.
func Build(ctx context.Context, cl *github.Client, req Request) (*Stats, error) {
	me := req.Login
	week := ResolveWeek(req.Now, req.Days)
	// GitHub search takes an ISO-8601 stamp with an offset, so the day
	// boundaries the chart draws are the ones the search filters on.
	since := week.Since.Format(time.RFC3339)

	queries := []github.Query{
		{Alias: "opened", Search: fmt.Sprintf("%s author:%s created:>=%s", searchBase, me, since)},
		{Alias: "merged", Search: fmt.Sprintf("%s author:%s merged:>=%s", searchBase, me, since)},
		{Alias: "closed", Search: fmt.Sprintf("%s author:%s is:unmerged closed:>=%s", searchBase, me, since)},
		{Alias: "reviewed", Search: fmt.Sprintf("%s reviewed-by:%s -author:%s updated:>=%s", searchBase, me, me, since)},
	}

	results, err := cl.BatchStatSearch(ctx, queries, req.Limit)
	if err != nil && len(results) == 0 {
		return nil, err
	}

	s := &Stats{
		Login:       me,
		Week:        week,
		GeneratedAt: req.Now,
		Daily:       emptyDays(week),
	}
	if err != nil {
		s.Warning = err.Error()
	}

	byDate := index(s.Daily)
	repos := map[string]int{}
	touch := func(repo string) {
		if repo != "" {
			repos[repo]++
		}
	}

	for _, pr := range results["opened"] {
		if !inWeek(pr.CreatedAt, week) {
			continue
		}
		s.Opened++
		touch(pr.Repository.NameWithOwner)
		if d := byDate[dayKey(pr.CreatedAt, week)]; d != nil {
			d.Opened++
		}
	}

	var fastest, biggest *github.PullRequestStat
	var fastestFor time.Duration
	for i, pr := range results["merged"] {
		if pr.MergedAt == nil || !inWeek(*pr.MergedAt, week) {
			continue
		}
		s.Merged++
		s.Additions += pr.Additions
		s.Deletions += pr.Deletions
		s.FilesChanged += pr.ChangedFiles
		touch(pr.Repository.NameWithOwner)
		if d := byDate[dayKey(*pr.MergedAt, week)]; d != nil {
			d.Merged++
		}
		// Superlatives are taken over merged pull requests only: an open
		// branch has no time-to-merge, and its diff is still moving.
		if took := pr.MergedAt.Sub(pr.CreatedAt); fastest == nil || took < fastestFor {
			fastest, fastestFor = &results["merged"][i], took
		}
		if lines := pr.Additions + pr.Deletions; biggest == nil || lines > biggest.Additions+biggest.Deletions {
			biggest = &results["merged"][i]
		}
	}

	for _, pr := range results["closed"] {
		if pr.MergedAt != nil || pr.ClosedAt == nil || !inWeek(*pr.ClosedAt, week) {
			continue
		}
		s.Closed++
		touch(pr.Repository.NameWithOwner)
	}

	for _, pr := range results["reviewed"] {
		written, approvals, days := countReviews(pr, me, week)
		if len(days) == 0 {
			continue
		}
		s.ReviewsWritten += written
		s.Approvals += approvals
		s.Reviewed++
		touch(pr.Repository.NameWithOwner)
		for _, key := range days {
			if d := byDate[key]; d != nil {
				d.Reviewed++
			}
		}
	}

	s.Repos = len(repos)
	s.Kcal = s.Opened*KcalPerOpened + s.Merged*KcalPerMerged + s.Closed*KcalPerClosed +
		s.ReviewsWritten*KcalPerReview + s.Approvals*KcalPerApproval
	s.ActiveDays, s.Streak = rhythm(s.Daily)

	if fastest != nil {
		s.Highlights.FastestMerge = ref(*fastest)
		s.Highlights.FastestMinutes = int(fastestFor / time.Minute)
	}
	if biggest != nil {
		s.Highlights.BiggestMerge = ref(*biggest)
		s.Highlights.BiggestLines = biggest.Additions + biggest.Deletions
	}
	s.Highlights.TopRepo, s.Highlights.TopRepoCount = topRepo(repos)
	return s, nil
}

// countReviews reduces one pull request from the reviewed search.
//
// The search only says "you have reviewed this at some point", so the review
// timestamps are what decide whether that happened inside the window. The
// returned days are the distinct columns this pull request contributes to:
// two rounds on the same branch on the same day are one column, not two.
func countReviews(pr github.PullRequestStat, me string, w Week) (written, approvals int, days []string) {
	seen := map[string]bool{}
	for _, r := range pr.Reviews.Nodes {
		if r.Author.LoginOr("") != me || !inWeek(r.CreatedAt, w) {
			continue
		}
		// PENDING is a review still being drafted in somebody's browser. It
		// has not been submitted, so it is not work delivered.
		if strings.EqualFold(r.State, "PENDING") {
			continue
		}
		written++
		if strings.EqualFold(r.State, "APPROVED") {
			approvals++
		}
		if key := dayKey(r.CreatedAt, w); !seen[key] {
			seen[key] = true
			days = append(days, key)
		}
	}
	sort.Strings(days)
	return written, approvals, days
}

func ref(pr github.PullRequestStat) *PRRef {
	return &PRRef{Repo: pr.Repository.NameWithOwner, Number: pr.Number, Title: pr.Title, URL: pr.URL}
}

// emptyDays lays out one Day per calendar day in the window, so a quiet day is
// a zero-height column rather than a gap in the chart.
func emptyDays(w Week) []Day {
	out := make([]Day, 0, w.Days)
	for i := 0; i < w.Days; i++ {
		out = append(out, Day{Date: w.Since.AddDate(0, 0, i).Format("2006-01-02")})
	}
	return out
}

func index(days []Day) map[string]*Day {
	out := make(map[string]*Day, len(days))
	for i := range days {
		out[days[i].Date] = &days[i]
	}
	return out
}

// dayKey buckets a timestamp by the local calendar day the window is drawn in,
// so a pull request merged at 23:50 lands on the column the user saw it on.
func dayKey(t time.Time, w Week) string {
	return t.In(w.Since.Location()).Format("2006-01-02")
}

func inWeek(t time.Time, w Week) bool {
	return !t.Before(w.Since) && !t.After(w.Until)
}

// rhythm counts the days with any activity, and the run of them ending now.
//
// Today is allowed to be empty without breaking the streak. The dashboard is
// something you open in the morning, and a counter that resets to zero every
// day until you have merged something is the opposite of encouraging.
func rhythm(days []Day) (active, streak int) {
	for _, d := range days {
		if d.Total() > 0 {
			active++
		}
	}
	i := len(days) - 1
	if i >= 0 && days[i].Total() == 0 {
		i--
	}
	for ; i >= 0 && days[i].Total() > 0; i-- {
		streak++
	}
	return active, streak
}

// topRepo is the repository the week happened in, ties broken by name so the
// answer does not shuffle between two equally busy repos on every refresh.
func topRepo(counts map[string]int) (string, int) {
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	best, bestN := "", 0
	for _, name := range names {
		if counts[name] > bestN {
			best, bestN = name, counts[name]
		}
	}
	return best, bestN
}
