// Package config resolves runtime configuration from the environment.
package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	// Addr is the listen address for the HTTP server.
	Addr string
	// StateDir holds settings.json and session.json.
	StateDir string
	// GraphQLEndpoint is the GitHub GraphQL endpoint.
	GraphQLEndpoint string
	// RESTEndpoint is the GitHub REST base, used for device flow discovery only.
	WebEndpoint string
	// ClientID enables the OAuth device flow when set.
	ClientID string
	// AllowGhCLI permits the backend to read the local gh CLI session.
	AllowGhCLI bool
	// EnvToken is a token handed in by the environment (docker/k8s bridge).
	EnvToken string
	// DevOrigin, when set, is allowed as a CORS origin for the vite dev server.
	DevOrigin string
	// AllowedHosts are extra Host header values accepted beyond loopback, for
	// deployments reached under a real hostname. See api.hostAllowed.
	AllowedHosts []string
	// DefaultLimit is the per-query PR cap.
	DefaultLimit int
}

func Load() Config {
	c := Config{
		Addr:            env("GHDASH_ADDR", "127.0.0.1:19080"),
		StateDir:        env("GHDASH_STATE_DIR", defaultStateDir()),
		GraphQLEndpoint: env("GHDASH_GRAPHQL_ENDPOINT", "https://api.github.com/graphql"),
		WebEndpoint:     env("GHDASH_GITHUB_WEB", "https://github.com"),
		ClientID:        env("GITHUB_CLIENT_ID", ""),
		AllowGhCLI:      envBool("GHDASH_ALLOW_GH_CLI", true),
		EnvToken:        firstNonEmpty(os.Getenv("GHDASH_GITHUB_TOKEN"), os.Getenv("GH_TOKEN"), os.Getenv("GITHUB_TOKEN")),
		DevOrigin:       env("GHDASH_DEV_ORIGIN", ""),
		AllowedHosts:    envList("GHDASH_ALLOWED_HOSTS"),
		DefaultLimit:    envInt("GHDASH_LIMIT", 25),
	}
	return c
}

func (c Config) SettingsPath() string { return filepath.Join(c.StateDir, "settings.json") }
func (c Config) SessionPath() string  { return filepath.Join(c.StateDir, "session.json") }

// defaultStateDir picks where settings.json and session.json live.
//
// The app was called github-dashboarder before it was called Yana-chan 4K, and
// the stored gh session is the expensive thing to lose -- losing it means
// authenticating again. So if the old directory is present and the new one is
// not, we keep reading the old one rather than silently starting fresh. Once
// the new directory exists it always wins.
func defaultStateDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ".yana"
	}
	current := filepath.Join(dir, "yana-chan-4k")
	if _, err := os.Stat(current); err == nil {
		return current
	}
	legacy := filepath.Join(dir, "github-dashboarder")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return current
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

// envList reads a comma-separated list, dropping empty entries.
func envList(key string) []string {
	var out []string
	for _, part := range strings.Split(os.Getenv(key), ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
