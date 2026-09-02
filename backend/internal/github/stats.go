package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// statBits is the shape the weekly stats need, and no more than that.
//
// It is deliberately not PRBits. The board asks "who said what, and when",
// so it drags along comments and review threads; the stats ask "what did I
// finish, and how big was it", which is closedAt/mergedAt, a diff size, the
// reviews the viewer themselves submitted, and who closed it. Sharing one
// fragment would mean every stats refresh paid for thirty comments per pull
// request that nothing ever reads.
//
// The timeline slice is the one field here that is not simply a column of the
// pull request. There is no `closedBy` on PullRequest and no `closed-by:`
// search qualifier, so the only way to know whether the viewer was the one who
// closed somebody else's branch is the event that closed it. `last:1` is the
// close that stuck, which is the one closedAt agrees with on a branch that was
// closed, reopened and closed again.
const statBits = `
fragment StatBits on PullRequest {
  number title url createdAt closedAt mergedAt
  additions deletions changedFiles
  repository { nameWithOwner }
  author { __typename login }
  reviews(last:50) { nodes { createdAt state author { __typename login } } }
  timelineItems(last:1, itemTypes:[CLOSED_EVENT]) {
    nodes { ... on ClosedEvent { createdAt actor { __typename login } } }
  }
}`

// PullRequestStat mirrors the StatBits fragment.
type PullRequestStat struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	URL          string     `json:"url"`
	CreatedAt    time.Time  `json:"createdAt"`
	ClosedAt     *time.Time `json:"closedAt"`
	MergedAt     *time.Time `json:"mergedAt"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
	ChangedFiles int        `json:"changedFiles"`
	Repository   Repository `json:"repository"`
	Author       *Actor     `json:"author"`
	Reviews      struct {
		Nodes []ReviewNode `json:"nodes"`
	} `json:"reviews"`
	TimelineItems struct {
		Nodes []ClosedEventNode `json:"nodes"`
	} `json:"timelineItems"`
}

// ClosedByLogin is who closed the pull request, or "" if nobody did or the
// account has since been deleted.
//
// A merge closes a branch too, and GitHub does not always file an event for
// that one, so this answers for the unmerged case only -- which is the case
// the week cares about.
func (p PullRequestStat) ClosedByLogin() string {
	if len(p.TimelineItems.Nodes) == 0 {
		return ""
	}
	return p.TimelineItems.Nodes[len(p.TimelineItems.Nodes)-1].Actor.LoginOr("")
}

// statSearchResult is one page of one alias.
type statSearchResult struct {
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
	Nodes []PullRequestStat `json:"nodes"`
}

// statPage is how many pull requests one page asks for: GraphQL's own
// per-connection maximum, so a normal week is still a single round trip.
const statPage = 100

// statMaxPages bounds one alias. Ninety days of a very busy account is the
// worst case, and an unbounded walk sits inside a request the browser is
// waiting on.
const statMaxPages = 5

// buildStatBatch is buildBatch with a cursor per alias, so each search can be
// carried on from where its last page stopped. The board has no use for this:
// a queue wants the top of the list, while the week has to count all of it.
func buildStatBatch(queries []Query, after map[string]string) (string, map[string]any) {
	params := []string{"$n:Int!"}
	fields := make([]string, 0, len(queries))
	vars := map[string]any{"n": statPage}
	for i, q := range queries {
		name, cursor := fmt.Sprintf("q%d", i), fmt.Sprintf("c%d", i)
		params = append(params, "$"+name+":String!", "$"+cursor+":String")
		fields = append(fields, fmt.Sprintf(
			"  %s: search(query:$%s, type:ISSUE, first:$n, after:$%s) { pageInfo { hasNextPage endCursor } nodes { ...StatBits } }",
			q.Alias, name, cursor))
		vars[name] = q.Search
		// A nil cursor is the first page; GraphQL takes `after: null` happily.
		if c := after[q.Alias]; c != "" {
			vars[cursor] = c
		} else {
			vars[cursor] = nil
		}
	}
	return fmt.Sprintf("query(%s) {\n%s\n}\n%s", strings.Join(params, ", "), strings.Join(fields, "\n"), statBits), vars
}

// BatchStatSearch runs every query over the lighter StatBits fragment and
// pages each one to the end of its window.
//
// Paging is the whole point. A single page would hand back whatever slice of
// the week GitHub felt like returning first, and the chart would then draw
// empty columns for days that were not empty at all -- the busier the week,
// the more of it went missing. Every page after the first only carries the
// aliases that said they had more, so the common week is one round trip and
// only a genuinely busy one costs a second.
//
// Same partial-failure contract as BatchSearch: the aliases that resolved come
// back alongside the error. An alias that fails on some page keeps what it had
// and stops there.
func (c *Client) BatchStatSearch(ctx context.Context, queries []Query) (map[string][]PullRequestStat, error) {
	out := map[string][]PullRequestStat{}
	if len(queries) == 0 {
		return out, nil
	}

	pending := queries
	cursors := map[string]string{}
	var failure error

	for page := 0; page < statMaxPages && len(pending) > 0; page++ {
		doc, vars := buildStatBatch(pending, cursors)

		var data map[string]statSearchResult
		err := c.Do(ctx, doc, vars, &data)
		if err != nil && failure == nil {
			failure = err
		}

		next := make([]Query, 0, len(pending))
		for _, q := range pending {
			res, ok := data[q.Alias]
			if !ok {
				// This alias did not resolve; failure says why.
				continue
			}
			out[q.Alias] = append(out[q.Alias], onlyPRs(res.Nodes)...)
			if res.PageInfo.HasNextPage && res.PageInfo.EndCursor != "" {
				cursors[q.Alias] = res.PageInfo.EndCursor
				next = append(next, q)
			}
		}
		pending = next
	}

	for alias, prs := range out {
		sort.SliceStable(prs, func(i, j int) bool { return prs[i].CreatedAt.After(prs[j].CreatedAt) })
		out[alias] = prs
	}
	return out, failure
}

// onlyPRs drops the nodes that are not pull requests: search returns a union,
// and an issue decodes into the struct as zero values.
func onlyPRs(nodes []PullRequestStat) []PullRequestStat {
	prs := make([]PullRequestStat, 0, len(nodes))
	for _, pr := range nodes {
		if pr.Number == 0 || pr.URL == "" {
			continue
		}
		prs = append(prs, pr)
	}
	return prs
}
