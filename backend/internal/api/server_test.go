package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christiannarbon/yanachan-4k/backend/internal/config"
	"github.com/christiannarbon/yanachan-4k/backend/internal/state"
)

// newTestHandler builds the real handler over a throwaway state directory.
// ui is nil, so only the API routes are registered.
func newTestHandler(t *testing.T, cfg config.Config) http.Handler {
	t.Helper()
	dir := t.TempDir()
	store, err := state.New(
		filepath.Join(dir, "settings.json"),
		filepath.Join(dir, "session.json"),
		state.DefaultSettings(25),
	)
	if err != nil {
		t.Fatalf("state.New: %v", err)
	}
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:19080"
	}
	return NewServer(cfg, store, nil).Handler()
}

func do(t *testing.T, h http.Handler, method, target, host string, headers map[string]string, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	r.Host = host
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// TestHostAllowlist is the DNS-rebinding guard. The attack that motivates it is
// the last case: evil.example resolves to 127.0.0.1, so Origin and Host agree
// and originAllowed alone would wave it through.
func TestHostAllowlist(t *testing.T) {
	h := newTestHandler(t, config.Config{Addr: "127.0.0.1:19080"})

	allowed := []string{
		"127.0.0.1:19080",
		"127.0.0.1",
		"127.0.0.2:19080", // all of 127.0.0.0/8 is loopback
		"localhost:19080",
		"LOCALHOST:19080", // case-insensitive
		"[::1]:19080",
		"app.localhost:19080",
	}
	for _, host := range allowed {
		if w := do(t, h, "GET", "/api/auth/status", host, nil, ""); w.Code == http.StatusForbidden {
			t.Errorf("Host %q: got 403, want it allowed", host)
		}
	}

	rejected := []string{
		"evil.example:19080",
		"rebind.attacker.test",
		"192.168.1.10:19080",
	}
	for _, host := range rejected {
		w := do(t, h, "GET", "/api/auth/status", host, nil, "")
		if w.Code != http.StatusForbidden {
			t.Errorf("Host %q: got %d, want 403", host, w.Code)
		}
	}
}

func TestHostAllowlistHonoursConfig(t *testing.T) {
	h := newTestHandler(t, config.Config{
		Addr:         "127.0.0.1:19080",
		AllowedHosts: []string{"dash.internal"},
	})
	if w := do(t, h, "GET", "/api/auth/status", "dash.internal:19080", nil, ""); w.Code == http.StatusForbidden {
		t.Errorf("configured host rejected: got %d", w.Code)
	}
	if w := do(t, h, "GET", "/api/auth/status", "other.internal:19080", nil, ""); w.Code != http.StatusForbidden {
		t.Errorf("unconfigured host: got %d, want 403", w.Code)
	}
}

// Kubernetes probes address the pod by IP, so health must answer to any Host.
func TestHealthExemptFromHostAllowlist(t *testing.T) {
	h := newTestHandler(t, config.Config{Addr: ":19080"})
	w := do(t, h, "GET", "/api/health", "10.244.1.7:19080", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("health from pod IP: got %d, want 200", w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := newTestHandler(t, config.Config{})
	w := do(t, h, "GET", "/api/auth/status", "127.0.0.1:19080", nil, "")

	want := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "no-referrer",
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	csp := w.Header().Get("Content-Security-Policy")
	// frame-ancestors is what stops a remote page framing the dashboard and
	// baiting a click onto the "approve my gh CLI token" button.
	for _, directive := range []string{"frame-ancestors 'none'", "default-src 'self'", "object-src 'none'"} {
		if !strings.Contains(csp, directive) {
			t.Errorf("CSP missing %q; got %q", directive, csp)
		}
	}
}

func TestOriginGuardOnMutatingRequests(t *testing.T) {
	const host = "127.0.0.1:19080"
	body := `{"teams":[],"orgs":[],"limit":25,"windowHours":0,"onlyActive":false,"showUrls":true}`

	cases := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{
			name:    "same origin accepted",
			headers: map[string]string{"Origin": "http://127.0.0.1:19080"},
			want:    http.StatusOK,
		},
		{
			name:    "cross origin rejected",
			headers: map[string]string{"Origin": "http://evil.example"},
			want:    http.StatusForbidden,
		},
		{
			name:    "no origin, non-browser client accepted",
			headers: nil,
			want:    http.StatusOK,
		},
		{
			name: "no origin but browser fetch metadata rejected",
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site",
			},
			want: http.StatusForbidden,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(t, config.Config{Addr: host})
			w := do(t, h, "PUT", "/api/settings", host, tc.headers, body)
			if w.Code != tc.want {
				t.Errorf("got %d, want %d (body %s)", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

// GET is not origin-guarded -- a cross-origin read is stopped by the browser,
// because we emit no CORS headers for anything but the dev origin.
func TestNoCORSHeadersWithoutDevOrigin(t *testing.T) {
	h := newTestHandler(t, config.Config{Addr: "127.0.0.1:19080"})
	w := do(t, h, "GET", "/api/auth/status", "127.0.0.1:19080",
		map[string]string{"Origin": "http://evil.example"}, "")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestDevOriginGetsCORS(t *testing.T) {
	const dev = "http://localhost:19090"
	h := newTestHandler(t, config.Config{Addr: "127.0.0.1:19080", DevOrigin: dev})
	w := do(t, h, "GET", "/api/auth/status", "127.0.0.1:19080",
		map[string]string{"Origin": dev}, "")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != dev {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, dev)
	}
}

// An API miss must answer in JSON. Falling through to the SPA returned
// index.html with a 200, which the frontend then surfaced as raw HTML.
func TestUnknownAPIRouteReturnsJSON(t *testing.T) {
	h := newTestHandler(t, config.Config{Addr: "127.0.0.1:19080"})
	w := do(t, h, "GET", "/api/nope", "127.0.0.1:19080", nil, "")

	if w.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}
	var payload map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("body is not JSON: %v (%s)", err, w.Body.String())
	}
	if payload["error"] == "" {
		t.Errorf("want an error message, got %v", payload)
	}
}

func TestSettingsValidation(t *testing.T) {
	const host = "127.0.0.1:19080"

	tooManyOrgs := make([]string, state.MaxRefs+1)
	for i := range tooManyOrgs {
		tooManyOrgs[i] = "org"
	}

	cases := []struct {
		name string
		body map[string]any
		want int
	}{
		{"valid", map[string]any{"teams": []string{"acme/platform"}, "orgs": []string{"acme"}, "limit": 25}, http.StatusOK},
		{"team without slash", map[string]any{"teams": []string{"platform"}, "orgs": []string{}}, http.StatusBadRequest},
		{"org with slash", map[string]any{"teams": []string{}, "orgs": []string{"acme/platform"}}, http.StatusBadRequest},
		{"too many orgs", map[string]any{"teams": []string{}, "orgs": tooManyOrgs}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(t, config.Config{Addr: host})
			b, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatal(err)
			}
			w := do(t, h, "PUT", "/api/settings", host, nil, string(b))
			if w.Code != tc.want {
				t.Errorf("got %d, want %d (%s)", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

// Every endpoint that touches GitHub must refuse before it has a session,
// rather than reaching for a nil client.
func TestUnauthenticatedEndpoints(t *testing.T) {
	h := newTestHandler(t, config.Config{Addr: "127.0.0.1:19080"})
	for _, path := range []string{"/api/board", "/api/stats", "/api/suggestions"} {
		w := do(t, h, "GET", path, "127.0.0.1:19080", nil, "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", path, w.Code)
		}
	}
}

// The token must never leave the backend. /api/auth/status is the one endpoint
// that reports on the session, so it is the one to watch.
func TestAuthStatusNeverLeaksToken(t *testing.T) {
	dir := t.TempDir()
	store, err := state.New(
		filepath.Join(dir, "settings.json"),
		filepath.Join(dir, "session.json"),
		state.DefaultSettings(25),
	)
	if err != nil {
		t.Fatal(err)
	}
	const secret = "ghp_averyrecognisablesecrettokenvalue"
	if err := store.SaveSession(state.Session{Mode: "gh-cli", Login: "octocat", Token: secret}); err != nil {
		t.Fatal(err)
	}
	h := NewServer(config.Config{Addr: "127.0.0.1:19080"}, store, nil).Handler()

	w := do(t, h, "GET", "/api/auth/status", "127.0.0.1:19080", nil, "")
	if strings.Contains(w.Body.String(), secret) {
		t.Fatalf("token leaked in /api/auth/status: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "octocat") {
		t.Errorf("want the login reported, got %s", w.Body.String())
	}
}
