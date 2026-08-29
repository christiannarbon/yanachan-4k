package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// prBits is the fragment from the original shell script, unchanged.
const prBits = `
fragment PRBits on PullRequest {
  number title url isDraft createdAt updatedAt reviewDecision
  repository { nameWithOwner }
  author { __typename login }
  commits(last:1) { nodes { commit { statusCheckRollup { state } } } }
  reviewRequests(first:20) {
    nodes { requestedReviewer { __typename ... on User { login } ... on Team { combinedSlug } } }
  }
  comments(last:30) { nodes { createdAt author { __typename login } } }
  reviews(last:30)  { nodes { createdAt state author { __typename login } } }
  reviewThreads(last:20) {
    nodes { isResolved comments(first:20) { nodes { createdAt author { __typename login } } } }
  }
}`

// Query is one aliased search in the batched GraphQL document.
type Query struct {
	Alias  string
	Search string
}

// BatchSearch runs every query in a single GraphQL request and returns the PRs
// keyed by alias. A partial failure (for example one unreadable org) still
// returns the aliases that did resolve, alongside the error.
func (c *Client) BatchSearch(ctx context.Context, queries []Query, limit int) (map[string][]PullRequest, error) {
	out := map[string][]PullRequest{}
	if len(queries) == 0 {
		return out, nil
	}

	var params []string
	var fields []string
	vars := map[string]any{"n": limit}
	params = append(params, "$n:Int!")
	for i, q := range queries {
		name := fmt.Sprintf("q%d", i)
		params = append(params, "$"+name+":String!")
		fields = append(fields, fmt.Sprintf("  %s: search(query:$%s, type:ISSUE, first:$n) { nodes { ...PRBits } }", q.Alias, name))
		vars[name] = q.Search
	}

	doc := fmt.Sprintf("query(%s) {\n%s\n}\n%s", strings.Join(params, ", "), strings.Join(fields, "\n"), prBits)

	var data map[string]SearchResult
	err := c.Do(ctx, doc, vars, &data)
	for alias, res := range data {
		prs := make([]PullRequest, 0, len(res.Nodes))
		for _, pr := range res.Nodes {
			// Search returns a union; non-PR nodes decode as zero values.
			if pr.Number == 0 || pr.URL == "" {
				continue
			}
			prs = append(prs, pr)
		}
		sort.SliceStable(prs, func(i, j int) bool { return prs[i].UpdatedAt.After(prs[j].UpdatedAt) })
		out[alias] = prs
	}
	return out, err
}
