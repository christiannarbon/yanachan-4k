package github

import "time"

type Actor struct {
	Typename string `json:"__typename"`
	Login    string `json:"login"`
}

func (a *Actor) LoginOr(fallback string) string {
	if a == nil || a.Login == "" {
		return fallback
	}
	return a.Login
}

type Repository struct {
	NameWithOwner string `json:"nameWithOwner"`
}

type StatusCheckRollup struct {
	State string `json:"state"`
}

type Commit struct {
	StatusCheckRollup *StatusCheckRollup `json:"statusCheckRollup"`
}

type CommitNode struct {
	Commit Commit `json:"commit"`
}

type RequestedReviewer struct {
	Typename     string `json:"__typename"`
	Login        string `json:"login"`
	CombinedSlug string `json:"combinedSlug"`
}

func (r RequestedReviewer) Slug() string {
	if r.CombinedSlug != "" {
		return r.CombinedSlug
	}
	return r.Login
}

type ReviewRequestNode struct {
	RequestedReviewer *RequestedReviewer `json:"requestedReviewer"`
}

type CommentNode struct {
	CreatedAt time.Time `json:"createdAt"`
	Author    *Actor    `json:"author"`
}

type ReviewNode struct {
	CreatedAt time.Time `json:"createdAt"`
	State     string    `json:"state"`
	Author    *Actor    `json:"author"`
}

type ReviewThreadNode struct {
	IsResolved bool `json:"isResolved"`
	Comments   struct {
		Nodes []CommentNode `json:"nodes"`
	} `json:"comments"`
}

// PullRequest mirrors the PRBits fragment.
type PullRequest struct {
	Number         int        `json:"number"`
	Title          string     `json:"title"`
	URL            string     `json:"url"`
	IsDraft        bool       `json:"isDraft"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	ReviewDecision string     `json:"reviewDecision"`
	Repository     Repository `json:"repository"`
	Author         *Actor     `json:"author"`
	Commits        struct {
		Nodes []CommitNode `json:"nodes"`
	} `json:"commits"`
	ReviewRequests struct {
		Nodes []ReviewRequestNode `json:"nodes"`
	} `json:"reviewRequests"`
	Comments struct {
		Nodes []CommentNode `json:"nodes"`
	} `json:"comments"`
	Reviews struct {
		Nodes []ReviewNode `json:"nodes"`
	} `json:"reviews"`
	ReviewThreads struct {
		Nodes []ReviewThreadNode `json:"nodes"`
	} `json:"reviewThreads"`
}

// ChecksState returns the rollup state of the newest commit, lowercased.
func (p PullRequest) ChecksState() string {
	if len(p.Commits.Nodes) == 0 {
		return ""
	}
	rollup := p.Commits.Nodes[0].Commit.StatusCheckRollup
	if rollup == nil {
		return ""
	}
	switch rollup.State {
	case "SUCCESS":
		return "success"
	case "FAILURE", "ERROR":
		return "failure"
	case "PENDING":
		return "pending"
	default:
		return ""
	}
}

type SearchResult struct {
	Nodes []PullRequest `json:"nodes"`
}
