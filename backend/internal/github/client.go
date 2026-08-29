// Package github is a small GraphQL client for the queries this dashboard needs.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	http     *http.Client
	endpoint string
	token    string
}

func New(endpoint, token string) *Client {
	return &Client{
		http:     &http.Client{Timeout: 45 * time.Second},
		endpoint: endpoint,
		token:    token,
	}
}

type gqlError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Do executes a GraphQL document and decodes the "data" object into out.
// GraphQL errors are returned even when data is partially populated, so
// callers that tolerate partial results should inspect the returned data first.
func (c *Client) Do(ctx context.Context, query string, vars map[string]any, out any) error {
	payload, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "yana-chan-4k")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github graphql: %s: %s", resp.Status, truncate(string(body), 300))
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []gqlError      `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode graphql response: %w", err)
	}
	if len(envelope.Data) > 0 && out != nil {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode graphql data: %w", err)
		}
	}
	if len(envelope.Errors) > 0 {
		msgs := make([]string, 0, len(envelope.Errors))
		for _, e := range envelope.Errors {
			msgs = append(msgs, e.Message)
		}
		return fmt.Errorf("github graphql: %s", strings.Join(msgs, "; "))
	}
	return nil
}

var ErrUnauthorized = errors.New("github rejected the token (401)")

// Viewer returns the login of the authenticated user.
func (c *Client) Viewer(ctx context.Context) (string, error) {
	var data struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}
	if err := c.Do(ctx, `query{viewer{login}}`, nil, &data); err != nil {
		return "", err
	}
	if data.Viewer.Login == "" {
		return "", errors.New("github returned an empty viewer login")
	}
	return data.Viewer.Login, nil
}

type Membership struct {
	Orgs  []string `json:"orgs"`
	Teams []string `json:"teams"`
}

// Memberships lists the orgs and teams the viewer belongs to, used to offer
// suggestions in the settings tab. Requires the read:org scope.
func (c *Client) Memberships(ctx context.Context) (Membership, error) {
	const q = `query {
  viewer {
    organizations(first: 100) {
      nodes {
        login
        teams(first: 100, role: MEMBER) { nodes { combinedSlug } }
      }
    }
  }
}`
	var data struct {
		Viewer struct {
			Organizations struct {
				Nodes []struct {
					Login string `json:"login"`
					Teams struct {
						Nodes []struct {
							CombinedSlug string `json:"combinedSlug"`
						} `json:"nodes"`
					} `json:"teams"`
				} `json:"nodes"`
			} `json:"organizations"`
		} `json:"viewer"`
	}
	err := c.Do(ctx, q, nil, &data)
	m := Membership{Orgs: []string{}, Teams: []string{}}
	for _, org := range data.Viewer.Organizations.Nodes {
		if org.Login != "" {
			m.Orgs = append(m.Orgs, org.Login)
		}
		for _, t := range org.Teams.Nodes {
			if t.CombinedSlug != "" {
				m.Teams = append(m.Teams, t.CombinedSlug)
			}
		}
	}
	return m, err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
