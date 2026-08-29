# Overview

Yana-chan 4K is a Go server and a Vue 3 single-page app that ship as **one
binary**. `make build-frontend` compiles the Vue bundle and stages it into
`backend/internal/webui/dist`; `go:embed` pulls it into the binary. Production is
one process, one port, no web server in front.

## The pieces

```
                       browser
                          │
                   http://127.0.0.1:19080
                          │
        ┌─────────────────▼──────────────────┐
        │  api.Server                        │
        │    host allowlist                  │
        │    security headers                │
        │    same-origin guard (non-GET)     │
        ├────────────────────────────────────┤
        │  /            → webui  (embedded)  │
        │  /api/auth/*  → authService        │
        │  /api/settings, /api/suggestions   │
        │  /api/board   → board.Build        │
        │  /api/stats   → stats.Build        │
        └───┬──────────────┬─────────────┬───┘
            │              │             │
       state.Store    github.Client   ghcli / ghauth
       settings.json   GraphQL,        gh auth token,
       session.json    batched         device flow
                          │
                  api.github.com
```

Everything is in-process. There is no database, no cache server and no
background worker: the two files in the state directory are the whole of the
persistent state, and every GitHub read happens inside the request that asked
for it.

## What a refresh looks like

The frontend issues **two** requests in parallel, because the landing tab should
not wait on the queues it is not showing:

1. `GET /api/stats` — the week.
2. `GET /api/board` — every queue tab.

Each of those becomes **one** GraphQL request to GitHub. Every tab's search is
an aliased field inside a single batched document, so following ten more teams
does not mean ten more round trips. See
[GraphQL batching](../low-level-design/graphql-batching.md).

The board request then flows:

```
handleBoard
  → store.Settings()          teams, orgs, limit, window override
  → board.ResolveWindow(now)  business-day rule, or the fixed override
  → github.BatchSearch(...)   one aliased document, one round trip
  → board.Build(...)          events → REPLY / NEW / quiet → sort → sections
  → JSON
```

`stats` is the same shape with a lighter fragment: the board needs comments and
review threads to answer "who said what", the week needs merge stamps and a diff
size, and would otherwise pay for thirty comments per pull request that nothing
reads.

## Dependency direction

```
cmd/server → api → { board, stats, state, config, github, ghcli, ghauth, webui }
                      board → github
                      stats → github
```

`board` and `stats` know about `github` and nothing else — no HTTP, no state, no
config. They are pure enough to test with fixtures, and both are tested that
way. Nothing under `internal/` imports `api`.

## Why it is shaped like this

- **One binary** because the thing this replaced was a shell script. A tool you
  run on your laptop to look at your own pull requests should not need a
  compose file to start.
- **No database** because there is nothing worth persisting except which teams
  you follow and one token. Both fit in a file, and a file can be `0600`.
- **Stateless reads** because GitHub is the source of truth and a cache would
  mean explaining, in the UI, how stale the number you are looking at is.
- **No LLM anywhere.** Not in the build, not at runtime.

## Where to go next

- [Backend](backend.md) — what each package owns.
- [Frontend](frontend.md) — components and the two pieces of global state.
- [Security model](security.md) — the token has no login in front of it.
