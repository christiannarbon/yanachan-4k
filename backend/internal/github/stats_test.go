package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// call is one GraphQL request the fake endpoint saw.
type call struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// fakeGitHub answers each request from reply, and records what it was asked.
func fakeGitHub(t *testing.T, reply func(page int, c call) string) (*Client, *[]call) {
	t.Helper()
	seen := &[]call{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var c call
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			t.Errorf("decode request: %v", err)
		}
		*seen = append(*seen, c)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, reply(len(*seen)-1, c))
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL, "token"), seen
}

func node(number int) string {
	return fmt.Sprintf(`{"number":%d,"url":"https://example.test/pr/%d","createdAt":"2026-08-24T09:00:00Z"}`, number, number)
}

func onePage(alias string, hasNext bool, cursor string, numbers ...int) string {
	nodes := make([]string, 0, len(numbers))
	for _, n := range numbers {
		nodes = append(nodes, node(n))
	}
	return fmt.Sprintf(`{"%s":{"pageInfo":{"hasNextPage":%t,"endCursor":"%s"},"nodes":[%s]}}`,
		alias, hasNext, cursor, strings.Join(nodes, ","))
}

// The bug this covers: a week busier than one page used to come back as one
// page, and the chart drew empty columns for days that were not empty.
func TestBatchStatSearchPagesToTheEndOfTheWindow(t *testing.T) {
	cl, seen := fakeGitHub(t, func(i int, c call) string {
		if i == 0 {
			return `{"data":` + onePage("opened", true, "CURSOR1", 1, 2) + `}`
		}
		return `{"data":` + onePage("opened", false, "", 3) + `}`
	})

	out, err := cl.BatchStatSearch(context.Background(), []Query{{Alias: "opened", Search: "is:pr author:me"}})
	if err != nil {
		t.Fatalf("BatchStatSearch: %v", err)
	}
	if got := len(out["opened"]); got != 3 {
		t.Fatalf("got %d pull requests, want all 3", got)
	}
	if len(*seen) != 2 {
		t.Fatalf("made %d requests, want 2", len(*seen))
	}
	if got := (*seen)[0].Variables["c0"]; got != nil {
		t.Errorf("first page asked after %v, want the start of the list", got)
	}
	if got := (*seen)[1].Variables["c0"]; got != "CURSOR1" {
		t.Errorf("second page asked after %v, want CURSOR1", got)
	}
}

// One page is one round trip: paging must not cost the quiet week anything.
func TestBatchStatSearchStopsWhenTheWindowIsExhausted(t *testing.T) {
	cl, seen := fakeGitHub(t, func(int, call) string {
		return `{"data":` + onePage("opened", false, "", 1) + `}`
	})

	if _, err := cl.BatchStatSearch(context.Background(), []Query{{Alias: "opened"}}); err != nil {
		t.Fatalf("BatchStatSearch: %v", err)
	}
	if len(*seen) != 1 {
		t.Errorf("made %d requests, want 1", len(*seen))
	}
}

// An endpoint that always says "more" must not be walked forever.
func TestBatchStatSearchStopsAtThePageCap(t *testing.T) {
	cl, seen := fakeGitHub(t, func(i int, _ call) string {
		return `{"data":` + onePage("opened", true, fmt.Sprintf("CURSOR%d", i), i+1) + `}`
	})

	if _, err := cl.BatchStatSearch(context.Background(), []Query{{Alias: "opened"}}); err != nil {
		t.Fatalf("BatchStatSearch: %v", err)
	}
	if len(*seen) != statMaxPages {
		t.Errorf("made %d requests, want the cap of %d", len(*seen), statMaxPages)
	}
}

// Only the aliases with more to give are carried onto the next page.
func TestBatchStatSearchDropsFinishedAliases(t *testing.T) {
	cl, seen := fakeGitHub(t, func(i int, _ call) string {
		if i == 0 {
			return `{"data":{"opened":{"pageInfo":{"hasNextPage":true,"endCursor":"C"},"nodes":[` + node(1) + `]},` +
				`"merged":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[` + node(2) + `]}}}`
		}
		return `{"data":` + onePage("opened", false, "", 3) + `}`
	})

	out, err := cl.BatchStatSearch(context.Background(), []Query{{Alias: "opened"}, {Alias: "merged"}})
	if err != nil {
		t.Fatalf("BatchStatSearch: %v", err)
	}
	if len(out["opened"]) != 2 || len(out["merged"]) != 1 {
		t.Errorf("opened = %d, merged = %d; want 2 and 1", len(out["opened"]), len(out["merged"]))
	}
	if strings.Contains((*seen)[1].Query, "merged:") {
		t.Error("second page still asked for merged, which had said it was done")
	}
}

// The partial-failure contract: what resolved comes back with the error.
func TestBatchStatSearchKeepsWhatResolved(t *testing.T) {
	cl, _ := fakeGitHub(t, func(int, call) string {
		return `{"data":{"opened":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[` + node(1) + `]}},` +
			`"errors":[{"message":"could not read org"}]}`
	})

	out, err := cl.BatchStatSearch(context.Background(), []Query{{Alias: "opened"}, {Alias: "merged"}})
	if err == nil {
		t.Fatal("want the GraphQL error alongside the results")
	}
	if len(out["opened"]) != 1 {
		t.Errorf("lost the alias that resolved: %v", out)
	}
	if _, ok := out["merged"]; ok {
		t.Error("merged did not resolve, so it should not be in the results")
	}
}

// Search returns a union: issues decode as zero values and must be dropped.
func TestBatchStatSearchDropsNonPullRequestNodes(t *testing.T) {
	cl, _ := fakeGitHub(t, func(int, call) string {
		return `{"data":{"opened":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[{},` + node(7) + `]}}}`
	})

	out, err := cl.BatchStatSearch(context.Background(), []Query{{Alias: "opened"}})
	if err != nil {
		t.Fatalf("BatchStatSearch: %v", err)
	}
	if len(out["opened"]) != 1 || out["opened"][0].Number != 7 {
		t.Errorf("got %v, want only the pull request", out["opened"])
	}
}
