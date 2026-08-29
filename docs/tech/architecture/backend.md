# Backend

Go 1.24, standard library only for HTTP. The toolchain is pinned to go1.25.0 in
`go.mod` so a local build, the Docker build and CI agree.

## Packages

```
backend/
  cmd/server/          entry point: config, store, server, graceful shutdown
  internal/api/        HTTP handlers, middleware, the auth service
  internal/board/      the ported classification logic, with tests
  internal/stats/      the weekly dashboard's reduction, with tests
  internal/github/     GraphQL client and the batched PR and stats searches
  internal/ghcli/      gh CLI detection
  internal/ghauth/     OAuth device flow
  internal/config/     environment → Config
  internal/state/      settings and session persistence
  internal/webui/      embeds the built frontend
```

### `cmd/server`

Loads config, opens the store, builds the handler, listens. Timeouts are set on
the `http.Server` rather than left at zero: 10s read-header, 90s write, 120s
idle. The write timeout is generous because a board refresh is a GraphQL call to
GitHub with a 45s client timeout of its own.

Shutdown is on `SIGINT`/`SIGTERM` with a 10s grace period.

### `internal/config`

One `Load()` reading the environment into a `Config` struct, with defaults. No
flags, no config file — every knob is an environment variable so that Docker and
Kubernetes need nothing extra. The full table is in
[configuration reference](../low-level-design/configuration-reference.md).

The one piece of logic here is `defaultStateDir()`, which prefers the current
`yana-chan-4k` directory but falls back to the pre-rename `github-dashboarder`
one if that is what exists. Losing the state directory means authenticating
again, so the rename does not do that to anyone.

### `internal/api`

Routing is `http.ServeMux` with Go 1.22 method patterns (`GET /api/board`). One
middleware chain wraps everything; see [the security model](security.md), which
is really a document about this file.

Two details that are easy to undo by accident:

- `/api/` has a catch-all returning JSON 404. Without it an unmatched API path
  falls through to the SPA handler, which answers `index.html` with a 200, and
  the frontend tries to `JSON.parse` a web page.
- Handlers write through `writeJSON` / `writeError` so every error the frontend
  sees has the same `{"error": "..."}` shape.

### `internal/board` and `internal/stats`

The two reductions. Both take GitHub types in and return a JSON payload; neither
imports `net/http`, `state` or `config`, which is what makes them testable with
fixtures and what keeps the HTTP layer thin.

`board` is a port of the jq program from the shell script this grew out of —
same event sources, same bot detection, same window, same `REPLY`/`NEW`
semantics, same sort order. `stats` is new. Details:
[board classification](../low-level-design/board-classification.md),
[weekly stats](../low-level-design/weekly-stats.md).

### `internal/github`

A small GraphQL client — `Do(ctx, query, vars, out)` — plus the two batched
searches. It deliberately returns GraphQL errors *alongside* partially populated
data, because one unreadable organization should not blank the other nine tabs.

### `internal/ghcli` and `internal/ghauth`

`ghcli` answers "is `gh` installed and logged in", without reading the token; the
token is only fetched once the user approves. `ghauth` is the OAuth device flow:
start, then poll, with the `slow_down` and `expired` states handled.

### `internal/state`

`Store` guards two JSON files with an `RWMutex`. Writes are atomic and `0600`.
Corrupt session files are tolerated — the user signs in again — while a corrupt
settings file is an error worth failing on. See
[state persistence](../low-level-design/state-persistence.md).

### `internal/webui`

`//go:embed all:dist` plus an SPA handler: real files when they exist,
`index.html` otherwise. Hashed assets under `assets/` get a one-year immutable
`Cache-Control`; `index.html` gets `no-store`, so a rebuilt bundle is picked up
on the next reload.

Only `dist/.gitkeep` is tracked. That is enough for `//go:embed all:dist` to
compile on a clean checkout -- the `all:` prefix matches dotfiles -- so the
backend builds before the frontend has ever been built. What it does not give
you is a page: until `make build-frontend` has staged a bundle, `/` answers
`frontend bundle is missing` with a 404.

## Tests

```sh
make test
```

- `board` — the window rule for every weekday, bot detection, each
  classification branch ported from the jq program.
- `stats` — the whole-day window, day buckets in a non-UTC zone, the streak
  rule, which reviews count.
- `api` — the Host allowlist, the origin guard, the security headers, the
  settings validator, and that the token does not appear in `/api/auth/status`.
- `state` — the `0600` file modes, the atomic write, tolerance for a corrupt
  session file.
