package board

import (
	"testing"
	"time"

	"github.com/christiannarbon/yanachan-4k/backend/internal/github"
)

func TestResolveWindow(t *testing.T) {
	cases := []struct {
		name      string
		day       time.Time
		override  int
		wantHours int
		wantKind  string
		wantLabel string
	}{
		{"monday goes back to friday", date(2026, 8, 24), 0, 72, WindowBusinessDay, "since last business day (Fri)"},
		{"tuesday is 24h", date(2026, 8, 25), 0, 24, WindowFixed, "last 24h"},
		{"friday is 24h", date(2026, 8, 28), 0, 24, WindowFixed, "last 24h"},
		{"saturday is 24h", date(2026, 8, 29), 0, 24, WindowBusinessDay, "since last business day (Fri)"},
		{"sunday is 48h", date(2026, 8, 30), 0, 48, WindowBusinessDay, "since last business day (Fri)"},
		{"override wins", date(2026, 8, 24), 6, 6, WindowFixed, "last 6h"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := ResolveWindow(tc.day, tc.override)
			if w.Hours != tc.wantHours {
				t.Errorf("hours = %d, want %d", w.Hours, tc.wantHours)
			}
			if w.Kind != tc.wantKind {
				t.Errorf("kind = %q, want %q", w.Kind, tc.wantKind)
			}
			if w.Label != tc.wantLabel {
				t.Errorf("label = %q, want %q", w.Label, tc.wantLabel)
			}
			if !w.Cutoff.Before(tc.day) {
				t.Errorf("cutoff %v is not before now %v", w.Cutoff, tc.day)
			}
		})
	}
}

func TestIsBot(t *testing.T) {
	bots := []github.Actor{
		{Typename: "Bot", Login: "whatever"},
		{Typename: "User", Login: "dependabot[bot]"},
		{Typename: "User", Login: "renovate"},
		{Typename: "User", Login: "github-actions"},
		{Typename: "User", Login: "my-release-bot"},
		{Typename: "User", Login: "CodeRabbitAI"},
	}
	for _, b := range bots {
		if !isBot(&b) {
			t.Errorf("%q should be a bot", b.Login)
		}
	}
	humans := []github.Actor{
		{Typename: "User", Login: "octocat"},
		{Typename: "User", Login: "robotham"},
		{Typename: "User", Login: "bottomley"},
	}
	for _, h := range humans {
		if isBot(&h) {
			t.Errorf("%q should not be a bot", h.Login)
		}
	}
}

func TestMineEntryFlagsHumanActivity(t *testing.T) {
	now := date(2026, 8, 28)
	cutoff := now.Add(-24 * time.Hour)
	pr := github.PullRequest{
		Number: 7, URL: "https://x/7", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now,
		Author: &github.Actor{Typename: "User", Login: "me"},
	}
	pr.Comments.Nodes = []github.CommentNode{
		{CreatedAt: now.Add(-2 * time.Hour), Author: &github.Actor{Typename: "User", Login: "alice"}},
		{CreatedAt: now.Add(-1 * time.Hour), Author: &github.Actor{Typename: "Bot", Login: "ci"}},
		{CreatedAt: now.Add(-90 * time.Hour), Author: &github.Actor{Typename: "User", Login: "bob"}},
		{CreatedAt: now.Add(-30 * time.Minute), Author: &github.Actor{Typename: "User", Login: "me"}},
	}
	e := mineEntry(pr, "me", cutoff)
	if !e.Active || !e.Hot {
		t.Fatalf("expected active and hot, got active=%v hot=%v", e.Active, e.Hot)
	}
	if e.Status != StatusNew {
		t.Errorf("status = %q, want %q", e.Status, StatusNew)
	}
	if e.HumanCount != 1 || e.BotCount != 1 {
		t.Errorf("humanCount=%d botCount=%d, want 1 and 1", e.HumanCount, e.BotCount)
	}
	if len(e.HumanActors) != 1 || e.HumanActors[0] != "alice" {
		t.Errorf("humanActors = %v, want [alice]", e.HumanActors)
	}
}

func TestMineEntryQuietWhenOnlyOldActivity(t *testing.T) {
	now := date(2026, 8, 28)
	cutoff := now.Add(-24 * time.Hour)
	pr := github.PullRequest{Number: 8, URL: "https://x/8", UpdatedAt: now, Author: &github.Actor{Login: "me"}}
	pr.Comments.Nodes = []github.CommentNode{
		{CreatedAt: now.Add(-50 * time.Hour), Author: &github.Actor{Typename: "User", Login: "alice"}},
	}
	e := mineEntry(pr, "me", cutoff)
	if e.Active || e.Hot {
		t.Fatalf("expected quiet entry, got active=%v hot=%v", e.Active, e.Hot)
	}
	if e.Status != StatusQuiet {
		t.Errorf("status = %q, want %q", e.Status, StatusQuiet)
	}
	if e.LastActivityAt == nil {
		t.Error("expected lastActivityAt to be carried over from the old comment")
	}
}

func TestReviewEntryDetectsReplyAfterYourComment(t *testing.T) {
	now := date(2026, 8, 28)
	cutoff := now.Add(-24 * time.Hour)
	pr := github.PullRequest{Number: 9, URL: "https://x/9", UpdatedAt: now, Author: &github.Actor{Login: "alice"}}
	pr.Reviews.Nodes = []github.ReviewNode{
		{CreatedAt: now.Add(-6 * time.Hour), State: "CHANGES_REQUESTED", Author: &github.Actor{Login: "me"}},
	}
	pr.ReviewThreads.Nodes = []github.ReviewThreadNode{{}}
	pr.ReviewThreads.Nodes[0].Comments.Nodes = []github.CommentNode{
		{CreatedAt: now.Add(-2 * time.Hour), Author: &github.Actor{Typename: "User", Login: "alice"}},
	}
	e := reviewEntry(pr, "me", "", cutoff)
	if e.Status != StatusReply {
		t.Fatalf("status = %q, want %q", e.Status, StatusReply)
	}
	if !e.Hot || !e.Active {
		t.Errorf("expected hot and active, got hot=%v active=%v", e.Hot, e.Active)
	}
	if e.YourState != "changes_requested" {
		t.Errorf("yourState = %q, want changes_requested", e.YourState)
	}
	if len(e.HumanActors) != 1 || e.HumanActors[0] != "alice" {
		t.Errorf("humanActors = %v, want [alice]", e.HumanActors)
	}
}

func TestReviewEntryAwaitingYouIsHot(t *testing.T) {
	now := date(2026, 8, 28)
	cutoff := now.Add(-24 * time.Hour)
	pr := github.PullRequest{Number: 10, URL: "https://x/10", UpdatedAt: now, Author: &github.Actor{Login: "alice"}}
	pr.ReviewRequests.Nodes = []github.ReviewRequestNode{
		{RequestedReviewer: &github.RequestedReviewer{Typename: "User", Login: "me"}},
	}
	e := reviewEntry(pr, "me", "", cutoff)
	if !e.Hot || !e.Active {
		t.Fatalf("an untouched PR requested from you must be hot and active")
	}
	if e.Awaiting != "you" {
		t.Errorf("awaiting = %q, want you", e.Awaiting)
	}
	if e.Touched {
		t.Error("touched should be false")
	}
}

func TestReviewEntryTeamModeAwaitsTeam(t *testing.T) {
	now := date(2026, 8, 28)
	cutoff := now.Add(-24 * time.Hour)
	pr := github.PullRequest{Number: 11, URL: "https://x/11", UpdatedAt: now, Author: &github.Actor{Login: "alice"}}
	pr.ReviewRequests.Nodes = []github.ReviewRequestNode{
		{RequestedReviewer: &github.RequestedReviewer{Typename: "Team", CombinedSlug: "acme/platform"}},
	}
	e := reviewEntry(pr, "me", "acme/platform", cutoff)
	if e.Awaiting != "team" {
		t.Errorf("awaiting = %q, want team", e.Awaiting)
	}
	if e.Hot {
		t.Error("a team-only request should be active but not hot")
	}
	if !e.Active {
		t.Error("a pending team request should be active")
	}
	if e.AlsoRequestedFromYou {
		t.Error("alsoRequestedFromYou should be false")
	}
}

func TestSortEntriesRanksHotThenActiveThenRecent(t *testing.T) {
	now := date(2026, 8, 28)
	in := []Entry{
		{Number: 1, UpdatedAt: now},
		{Number: 2, Active: true, UpdatedAt: now.Add(-10 * time.Hour)},
		{Number: 3, Hot: true, Active: true, UpdatedAt: now.Add(-20 * time.Hour)},
		{Number: 4, Active: true, UpdatedAt: now.Add(-1 * time.Hour)},
	}
	sortEntries(in)
	got := []int{in[0].Number, in[1].Number, in[2].Number, in[3].Number}
	want := []int{3, 4, 2, 1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestMakeSectionCountsBeforeFiltering(t *testing.T) {
	entries := []Entry{
		{Number: 1, Hot: true, Active: true},
		{Number: 2, Active: true},
		{Number: 3},
	}
	s := makeSection("mine", "Your open PRs", KindMine, "", entries, true)
	if s.Total != 3 || s.Active != 2 || s.Hot != 1 {
		t.Errorf("counts total=%d active=%d hot=%d, want 3/2/1", s.Total, s.Active, s.Hot)
	}
	if len(s.Entries) != 2 {
		t.Errorf("onlyActive should keep 2 entries, kept %d", len(s.Entries))
	}
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
}
