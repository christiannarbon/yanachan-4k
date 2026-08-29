// Package webui serves the compiled Vue application.
//
// dist/ is populated by `make build-frontend`, which copies frontend/dist here
// before the Go binary is compiled. The placeholder index.html keeps `go build`
// working on a clean checkout.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var assets embed.FS

// Handler serves the SPA: real files when they exist, index.html otherwise.
func Handler() http.Handler {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" || clean == "." {
			serveIndex(w, r, sub)
			return
		}
		if _, err := fs.Stat(sub, clean); err != nil {
			serveIndex(w, r, sub)
			return
		}
		if strings.HasPrefix(clean, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, sub fs.FS) {
	b, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		http.Error(w, "frontend bundle is missing; run 'make build-frontend'", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}
