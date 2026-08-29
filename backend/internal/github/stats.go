package github

import (
	"context"
	"sort"
	"time"
)

// statBits is the shape the weekly stats need, and no more than that.
//
// It is deliberately not PRBits. The board asks "who said what, and when",
// so it drags along comments and review threads; the stats ask "what did I
// finish, and how big was it", which is closedAt/mergedAt, a diff size, and
// the reviews the viewer themselves submitted. Sharing one fragment would
// mean every stats refresh paid for thirty comments per pull request that
// nothing ever reads.
const statBits = `
fragment StatBits on PullRequest {
  number title url createdAt closedAt mergedAt
  additions deletions changedFiles
  repository { nameWithOwner }
  author { __typename login }
  reviews(last:50) { nodes { createdAt state author { __typename login } } }
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
}

type statSearchResult struct {
	Nodes []PullRequestStat `json:"nodes"`
}

// BatchStatSearch runs every query in a single GraphQL request over the
// lighter StatBits fragment, with the same partial-failure contract as
// BatchSearch: the aliases that resolved come back alongside the error.
func (c *Client) BatchStatSearch(ctx context.Context, queries []Query, limit int) (map[string][]PullRequestStat, error) {
	out := map[string][]PullRequestStat{}
	if len(queries) == 0 {
		return out, nil
	}

	doc, vars := buildBatch(queries, statBits, "StatBits", limit)

	var data map[string]statSearchResult
	err := c.Do(ctx, doc, vars, &data)
	for alias, res := range data {
		prs := make([]PullRequestStat, 0, len(res.Nodes))
		for _, pr := range res.Nodes {
			// Search returns a union; non-PR nodes decode as zero values.
			if pr.Number == 0 || pr.URL == "" {
				continue
			}
			prs = append(prs, pr)
		}
		sort.SliceStable(prs, func(i, j int) bool { return prs[i].CreatedAt.After(prs[j].CreatedAt) })
		out[alias] = prs
	}
	return out, err
}
