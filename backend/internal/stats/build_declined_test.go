package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/christiannarbon/yanachan-4k/backend/internal/github"
)

// fakeGitHub answers every alias in the batch from nodes, keyed by alias, and
// records the search strings it was asked for.
func fakeGitHub(t *testing.T, nodes map[string]string) (*github.Client, *map[string]string) {
	t.Helper()
	asked := &map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		// The batch names each alias in a `alias: search(query:$qN` field, so
		// the document says which variable carried which alias's search.
		out := make([]string, 0, len(nodes))
		for _, line := range strings.Split(req.Query, "\n") {
			alias, rest, ok := strings.Cut(strings.TrimSpace(line), ": search(query:$")
			if !ok {
				continue
			}
			name, _, _ := strings.Cut(rest, ",")
			(*asked)[alias], _ = req.Variables[name].(string)
			out = append(out, fmt.Sprintf(`"%s":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[%s]}`,
				alias, nodes[alias]))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{%s}}`, strings.Join(out, ","))
	}))
	t.Cleanup(srv.Close)
	return github.New(srv.URL, "token"), asked
}

// closedNode is one node off the declined search, as GitHub sends it.
func closedNode(number int, repo, author, closer string, when time.Time) string {
	return fmt.Sprintf(`{"number":%d,"url":"https://example.test/pr/%d","createdAt":"%s","closedAt":"%s",
		"repository":{"nameWithOwner":"%s"},"author":{"__typename":"Bot","login":"%s"},
		"timelineItems":{"nodes":[{"createdAt":"%s","actor":{"__typename":"User","login":"%s"}}]}}`,
		number, number, when.Format(time.RFC3339), when.Format(time.RFC3339), repo, author,
		when.Format(time.RFC3339), closer)
}

// The bug this covers: closing a bot's dependency bump is triage that leaves
// no trace on your own account, so a morning of it used to read as an empty
// week. Reported as issue #24.
func TestBuildCountsTheBranchesYouClosedForSomebodyElse(t *testing.T) {
	now := at(2026, time.August, 30, 14)
	when := at(2026, time.August, 29, 10)

	cl, asked := fakeGitHub(t, map[string]string{
		"declined": strings.Join([]string{
			closedNode(1, "acme/api", "dependabot", "me", when),
			closedNode(2, "acme/api", "renovate", "me", when),
			// Superseded by the bot itself: their afternoon, not yours.
			closedNode(3, "acme/web", "dependabot", "dependabot", when),
			// A colleague closed one you had merely commented on.
			closedNode(4, "acme/web", "renovate", "alice", when),
			// Yours to close, but not this week.
			closedNode(5, "acme/web", "dependabot", "me", at(2026, time.August, 20, 10)),
		}, ","),
	})

	s, err := Build(context.Background(), cl, Request{Login: "me", Days: 7, Now: now})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if s.Closed != 2 {
		t.Errorf("closed = %d, want the two you closed yourself", s.Closed)
	}
	// The repository you spent the morning triaging is a repository you
	// touched, even though nothing there carries your name.
	if s.Repos != 1 {
		t.Errorf("repos = %d, want 1", s.Repos)
	}
	if want := 2 * KcalPerClosed; s.Kcal != want {
		t.Errorf("kcal = %d, want %d", s.Kcal, want)
	}

	q := (*asked)["declined"]
	for _, part := range []string{"is:pr", "-author:me", "is:unmerged", "is:closed", "involves:me", "closed:>=2026-08-24T00:00:00+09:00"} {
		if !strings.Contains(q, part) {
			t.Errorf("declined search %q is missing %q", q, part)
		}
	}
}

// The two closed searches must not both count the same branch. They cannot:
// one is author:me and the other -author:me. This pins that down, because
// widening either one would double the tile without any test noticing.
func TestBuildDoesNotCountAClosedBranchTwice(t *testing.T) {
	now := at(2026, time.August, 30, 14)
	when := at(2026, time.August, 29, 10)

	cl, _ := fakeGitHub(t, map[string]string{
		"closed":   closedNode(7, "acme/api", "me", "alice", when),
		"declined": closedNode(8, "acme/api", "dependabot", "me", when),
	})

	s, err := Build(context.Background(), cl, Request{Login: "me", Days: 7, Now: now})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if s.Closed != 2 {
		t.Errorf("closed = %d, want one of yours and one of theirs", s.Closed)
	}
}
