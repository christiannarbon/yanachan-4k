// Package ghcli inspects the gh CLI installed on the host machine.
//
// Nothing here is called until the user has explicitly approved it, except
// Detect, which only reports whether gh exists and whether it is logged in.
// Detect never reads the token.
package ghcli

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Status struct {
	Available     bool   `json:"available"`
	Authenticated bool   `json:"authenticated"`
	Login         string `json:"login"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	Detail        string `json:"detail,omitempty"`
}

// "Logged in to github.com account octocat (keyring)"
var loginRe = regexp.MustCompile(`(?i)logged in to ([^\s]+) (?:account|as) ([^\s]+)`)

func Detect(ctx context.Context) Status {
	var st Status
	path, err := exec.LookPath("gh")
	if err != nil {
		st.Detail = "gh CLI was not found on PATH"
		return st
	}
	st.Available = true
	st.Path = path

	out, err := run(ctx, "gh", "auth", "status")
	if err != nil {
		st.Detail = strings.TrimSpace(firstLine(out))
		if st.Detail == "" {
			st.Detail = "gh is installed but not authenticated"
		}
		return st
	}
	st.Authenticated = true
	if m := loginRe.FindStringSubmatch(out); m != nil {
		st.Host = m[1]
		st.Login = strings.Trim(m[2], "()")
	}
	if st.Host == "" {
		st.Host = "github.com"
	}
	return st
}

// Token returns the token backing the local gh session. Only call this after
// the user has approved the gh CLI auth mode.
func Token(ctx context.Context) (string, error) {
	out, err := run(ctx, "gh", "auth", "token")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func run(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	// gh auth status writes its report to stderr on some versions.
	combined := stdout.String()
	if strings.TrimSpace(combined) == "" {
		combined = stderr.String()
	}
	return combined, err
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}
