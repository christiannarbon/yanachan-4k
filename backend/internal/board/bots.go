package board

import (
	"regexp"

	"github.com/christiannarbon/yanachan-4k/backend/internal/github"
)

// botLogin matches the same set of automation accounts as the shell script.
var botLogin = regexp.MustCompile(`(?i)\[bot\]|-bot$|^bot$|^dependabot|^renovate|^copilot|^github-actions|^coderabbit|^codecov|^sonar|^snyk|^netlify|^vercel|^mergify|^stale`)

func isBot(a *github.Actor) bool {
	if a == nil {
		return false
	}
	if a.Typename == "Bot" {
		return true
	}
	return botLogin.MatchString(a.Login)
}
