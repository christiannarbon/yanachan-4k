// Package board turns GitHub search results into the dashboard payload.
//
// The classification rules are a direct port of the jq program in the shell
// script this dashboard grew out of: same event sources, same bot detection,
// same activity window, same REPLY / NEW / quiet semantics and the same sort
// order.
package board

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/christiannarbon/yanachan-4k/backend/internal/github"
)

type Request struct {
	Login      string
	AuthMode   string
	Teams      []string
	Orgs       []string
	Limit      int
	Window     Window
	OnlyActive bool
}

type event struct {
	who   string
	bot   bool
	at    time.Time
	kind  string
	state string
}

// events flattens issue comments, reviews and review-thread comments into one
// timeline, exactly like the jq "evs" function.
func events(pr github.PullRequest) []event {
	var out []event
	for _, c := range pr.Comments.Nodes {
		out = append(out, event{who: c.Author.LoginOr("ghost"), bot: isBot(c.Author), at: c.CreatedAt, kind: "comment"})
	}
	for _, r := range pr.Reviews.Nodes {
		out = append(out, event{who: r.Author.LoginOr("ghost"), bot: isBot(r.Author), at: r.CreatedAt, kind: "review", state: r.State})
	}
	for _, t := range pr.ReviewThreads.Nodes {
		for _, c := range t.Comments.Nodes {
			out = append(out, event{who: c.Author.LoginOr("ghost"), bot: isBot(c.Author), at: c.CreatedAt, kind: "thread"})
		}
	}
	return out
}

func filterEvents(in []event, keep func(event) bool) []event {
	out := make([]event, 0, len(in))
	for _, e := range in {
		if keep(e) {
			out = append(out, e)
		}
	}
	return out
}

func maxAt(in []event) time.Time {
	var max time.Time
	for _, e := range in {
		if e.at.After(max) {
			max = e.at
		}
	}
	return max
}

func actors(in []event) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, e := range in {
		if seen[e.who] {
			continue
		}
		seen[e.who] = true
		out = append(out, e.who)
	}
	sort.Strings(out)
	return out
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	c := t
	return &c
}

func base(pr github.PullRequest) Entry {
	return Entry{
		Number:         pr.Number,
		Title:          pr.Title,
		URL:            pr.URL,
		Repo:           pr.Repository.NameWithOwner,
		IsDraft:        pr.IsDraft,
		Author:         pr.Author.LoginOr("ghost"),
		AuthorIsBot:    isBot(pr.Author),
		CreatedAt:      pr.CreatedAt,
		UpdatedAt:      pr.UpdatedAt,
		Checks:         pr.ChecksState(),
		ReviewDecision: strings.ToLower(pr.ReviewDecision),
		HumanActors:    []string{},
		BotActors:      []string{},
	}
}

// mineEntry classifies one of the viewer's own pull requests: did anybody else
// touch it inside the window?
func mineEntry(pr github.PullRequest, me string, cutoff time.Time) Entry {
	e := base(pr)
	all := events(pr)
	others := filterEvents(all, func(ev event) bool { return ev.who != me })
	recent := filterEvents(others, func(ev event) bool { return ev.at.After(cutoff) })
	humans := filterEvents(recent, func(ev event) bool { return !ev.bot })
	bots := filterEvents(recent, func(ev event) bool { return ev.bot })

	e.HumanCount = len(humans)
	e.BotCount = len(bots)
	e.HumanActors = actors(humans)
	e.BotActors = actors(bots)
	e.Active = len(recent) > 0
	e.Hot = len(humans) > 0
	if e.Active {
		e.Status = StatusNew
		e.LastActivityAt = timePtr(maxAt(recent))
	} else {
		e.Status = StatusQuiet
		e.LastActivityAt = timePtr(maxAt(others))
	}
	return e
}

// reviewEntry classifies a pull request sitting in the viewer's review queue.
// teamSlug is non-empty for a team tab, which changes the "awaiting" wording.
func reviewEntry(pr github.PullRequest, me, teamSlug string, cutoff time.Time) Entry {
	e := base(pr)
	all := events(pr)
	mine := filterEvents(all, func(ev event) bool { return ev.who == me })
	others := filterEvents(all, func(ev event) bool { return ev.who != me })
	myLast := maxAt(mine)

	myState := ""
	var myStateAt time.Time
	for _, ev := range mine {
		if ev.kind == "review" && !ev.at.Before(myStateAt) {
			myStateAt = ev.at
			myState = ev.state
		}
	}

	replies := filterEvents(others, func(ev event) bool { return ev.at.After(myLast) && ev.at.After(cutoff) })
	recent := filterEvents(others, func(ev event) bool { return ev.at.After(cutoff) })
	replyHumans := filterEvents(replies, func(ev event) bool { return !ev.bot })
	replyBots := filterEvents(replies, func(ev event) bool { return ev.bot })

	var pendingMe, pendingTeam bool
	for _, rr := range pr.ReviewRequests.Nodes {
		if rr.RequestedReviewer == nil {
			continue
		}
		slug := rr.RequestedReviewer.Slug()
		if strings.EqualFold(slug, me) {
			pendingMe = true
		}
		if teamSlug != "" && strings.EqualFold(slug, teamSlug) {
			pendingTeam = true
		}
	}
	touched := len(mine) > 0

	e.Touched = touched
	e.Active = len(replies) > 0 || (!touched && (len(recent) > 0 || pendingMe || pendingTeam))
	e.Hot = len(replyHumans) > 0 || (!touched && pendingMe)

	switch {
	case len(replies) > 0:
		e.Status = StatusReply
		e.HumanCount = len(replyHumans)
		e.BotCount = len(replyBots)
		e.HumanActors = actors(replyHumans)
		e.BotActors = actors(replyBots)
		e.LastActivityAt = timePtr(maxAt(replies))
	case !touched && len(recent) > 0:
		e.Status = StatusNew
		humans := filterEvents(recent, func(ev event) bool { return !ev.bot })
		bots := filterEvents(recent, func(ev event) bool { return ev.bot })
		e.HumanCount = len(humans)
		e.BotCount = len(bots)
		e.HumanActors = actors(humans)
		e.BotActors = actors(bots)
		e.LastActivityAt = timePtr(maxAt(recent))
	default:
		e.Status = StatusQuiet
		e.LastActivityAt = timePtr(maxAt(others))
	}

	if touched {
		switch myState {
		case "APPROVED":
			e.YourState = "approved"
		case "CHANGES_REQUESTED":
			e.YourState = "changes_requested"
		default:
			e.YourState = "commented"
		}
		e.YourLastAt = timePtr(myLast)
	} else if teamSlug != "" && !pendingMe {
		e.Awaiting = "team"
	} else {
		e.Awaiting = "you"
	}
	e.AlsoRequestedFromYou = teamSlug != "" && pendingMe
	return e
}

// sortEntries ranks hot before active before quiet, then most recently updated.
func sortEntries(in []Entry) {
	rank := func(e Entry) int {
		switch {
		case e.Hot:
			return 0
		case e.Active:
			return 1
		default:
			return 2
		}
	}
	sort.SliceStable(in, func(i, j int) bool {
		if r1, r2 := rank(in[i]), rank(in[j]); r1 != r2 {
			return r1 < r2
		}
		return in[i].UpdatedAt.After(in[j].UpdatedAt)
	})
}

func makeSection(id, title, kind, ref string, entries []Entry, onlyActive bool) Section {
	s := Section{ID: id, Title: title, Kind: kind, Ref: ref, Entries: []Entry{}}
	for _, e := range entries {
		s.Total++
		if e.Active {
			s.Active++
		}
		if e.Hot {
			s.Hot++
		}
		if onlyActive && !e.Active {
			continue
		}
		s.Entries = append(s.Entries, e)
	}
	sortEntries(s.Entries)
	return s
}

const searchBase = "is:open is:pr archived:false"

// Build fetches every configured search in one GraphQL round trip and assembles
// the tabbed board.
func Build(ctx context.Context, cl *github.Client, req Request) (*Board, error) {
	me := req.Login
	queries := []github.Query{
		{Alias: "mine", Search: fmt.Sprintf("%s author:%s sort:updated-desc", searchBase, me)},
		{Alias: "req", Search: fmt.Sprintf("%s user-review-requested:%s sort:updated-desc", searchBase, me)},
		{Alias: "revd", Search: fmt.Sprintf("%s reviewed-by:%s -author:%s sort:updated-desc", searchBase, me, me)},
	}
	for i, team := range req.Teams {
		queries = append(queries, github.Query{
			Alias:  fmt.Sprintf("team%d", i),
			Search: fmt.Sprintf("%s team-review-requested:%s sort:updated-desc", searchBase, team),
		})
	}
	for i, org := range req.Orgs {
		queries = append(queries, github.Query{
			Alias:  fmt.Sprintf("org%d", i),
			Search: fmt.Sprintf("%s org:%s involves:%s sort:updated-desc", searchBase, org, me),
		})
	}

	results, err := cl.BatchSearch(ctx, queries, req.Limit)
	if err != nil && len(results) == 0 {
		return nil, err
	}

	b := &Board{
		Login:       me,
		AuthMode:    req.AuthMode,
		Window:      req.Window,
		GeneratedAt: time.Now(),
		OnlyActive:  req.OnlyActive,
		Limit:       req.Limit,
	}
	if err != nil {
		b.Warning = err.Error()
	}
	cutoff := req.Window.Cutoff

	// Section 1: the viewer's own open PRs.
	mineEntries := make([]Entry, 0, len(results["mine"]))
	for _, pr := range results["mine"] {
		mineEntries = append(mineEntries, mineEntry(pr, me, cutoff))
	}

	// Team sections are built first: the script gives a team-requested PR to
	// the team section and drops it from the personal review queue.
	teamURLs := map[string]bool{}
	teamSections := make([]Section, 0, len(req.Teams))
	for i, team := range req.Teams {
		prs := results[fmt.Sprintf("team%d", i)]
		entries := make([]Entry, 0, len(prs))
		for _, pr := range prs {
			teamURLs[pr.URL] = true
			entries = append(entries, reviewEntry(pr, me, team, cutoff))
		}
		teamSections = append(teamSections, makeSection("team:"+team, team, KindTeam, team, entries, req.OnlyActive))
	}

	// Section 2: review requested from you, plus PRs you already reviewed.
	seen := map[string]bool{}
	reviewEntries := []Entry{}
	for _, alias := range []string{"req", "revd"} {
		for _, pr := range results[alias] {
			if seen[pr.URL] || teamURLs[pr.URL] {
				continue
			}
			seen[pr.URL] = true
			reviewEntries = append(reviewEntries, reviewEntry(pr, me, "", cutoff))
		}
	}

	// Org sections are additive, so they exclude anything already on a tab.
	shown := map[string]bool{}
	for _, e := range mineEntries {
		shown[e.URL] = true
	}
	for url := range teamURLs {
		shown[url] = true
	}
	for _, e := range reviewEntries {
		shown[e.URL] = true
	}
	orgSections := make([]Section, 0, len(req.Orgs))
	for i, org := range req.Orgs {
		prs := results[fmt.Sprintf("org%d", i)]
		entries := make([]Entry, 0, len(prs))
		for _, pr := range prs {
			if shown[pr.URL] {
				continue
			}
			if pr.Author.LoginOr("") == me {
				entries = append(entries, mineEntry(pr, me, cutoff))
				continue
			}
			entries = append(entries, reviewEntry(pr, me, "", cutoff))
		}
		orgSections = append(orgSections, makeSection("org:"+org, org, KindOrg, org, entries, req.OnlyActive))
	}

	b.Sections = append(b.Sections,
		makeSection("mine", "Your open PRs", KindMine, "", mineEntries, req.OnlyActive),
		makeSection("review", "Review requested from you", KindReview, "", reviewEntries, req.OnlyActive),
	)
	b.Sections = append(b.Sections, teamSections...)
	b.Sections = append(b.Sections, orgSections...)
	return b, nil
}
