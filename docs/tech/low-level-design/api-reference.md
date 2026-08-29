# API reference

`backend/internal/api/server.go`

All JSON. Routing is `http.ServeMux` with method patterns, so a wrong method on
a real path is a 405 rather than a 404.

## Conventions

- Errors are `{"error": "..."}` with a status code, written by `writeError`.
- Every non-`GET`/`HEAD` request passes the [same-origin guard](../architecture/security.md#guard-2--mutating-requests-must-be-same-origin).
- Every request except `/api/health` passes the [Host allowlist](../architecture/security.md#guard-3--the-host-header-is-pinned-to-loopback).
- An unmatched `/api/` path returns a JSON 404, not the SPA's `index.html`.
- **No response ever contains the token.**

## Health

### `GET /api/health`

```json
{ "status": "ok", "time": "2026-08-30T09:12:00+09:00" }
```

Exempt from the Host allowlist so Kubernetes probes can address the pod by IP.
Requires no session.

## Authentication

### `GET /api/auth/status`

What sign-in paths exist, and whether one has been taken.

```json
{
  "authenticated": true,
  "session": { "mode": "gh-cli", "login": "octocat" },
  "ghCli": { "available": true, "authenticated": true, "login": "octocat",
             "host": "github.com", "path": "/opt/homebrew/bin/gh" },
  "ghCliAllowed": true,
  "oauthEnabled": false,
  "envTokenAvailable": false,
  "oauthScopes": "repo read:org",
  "pendingDevice": null
}
```

`session` is null when unauthenticated. `pendingDevice` appears only while a
device flow is in progress and carries `userCode`, `verificationUri`,
`expiresAt` and `interval`.

`TestAuthStatusNeverLeaksToken` asserts the token string does not appear
anywhere in this body. Adding a field here means re-checking that test.

### `POST /api/auth/gh-cli/approve`

Runs `gh auth token` — this is the first moment the token is read — and stores
the session.

`200 → {"mode": "gh-cli", "login": "octocat"}`, `400` if `gh` is missing, not
logged in, or disallowed by `GHDASH_ALLOW_GH_CLI=false`.

### `POST /api/auth/env-token/approve`

Same, for a token already in the environment. `400` when there is none.

### `POST /api/auth/device/start`

```json
{ "userCode": "ABCD-1234", "verificationUri": "https://github.com/login/device",
  "expiresIn": 900, "interval": 5 }
```

`400` when `GITHUB_CLIENT_ID` is unset.

### `POST /api/auth/device/poll`

```json
{ "state": "pending", "session": null }
```

`state` is one of `pending`, `slow_down`, `complete`, `expired`, `denied`.
`session` is present only on `complete`. The frontend polls at `interval`
seconds and backs off on `slow_down`.

### `POST /api/auth/logout`

Deletes `session.json`. Settings survive.

## Settings

### `GET /api/settings`

The normalized settings — what the store holds, not what was last sent.

```json
{ "teams": ["acme/platform"], "orgs": ["acme"], "limit": 25,
  "windowHours": 0, "onlyActive": false, "showUrls": true }
```

### `PUT /api/settings`

Body is the same shape. Validation, in order:

| Check | Response |
| --- | --- |
| Body over 1 MiB or not JSON | `400 invalid settings payload: ...` |
| More than 200 teams or 200 orgs | `400 at most 200 teams and 200 organizations...` |
| A team without a `/` | `400 team must be in org/team-slug form: X` |
| An org with a `/` | `400 org must be a bare organization login: X` |

Anything that passes is then [normalized](state-persistence.md#normalization)
rather than rejected: out-of-range numbers are corrected, entries are trimmed
and deduped. The response is the stored result, so the UI can show what actually
took effect.

### `GET /api/suggestions`

The orgs and teams the viewer belongs to, for the Settings tab's one-click
suggestions.

```json
{ "orgs": ["acme"], "teams": ["acme/platform"], "warning": "..." }
```

`401` without a session. `502` only when the memberships query failed **and**
returned nothing; a partial result comes back `200` with `warning` set —
typically a token without `read:org`.

## The dashboard

### `GET /api/board`

| Parameter | Range | Default |
| --- | --- | --- |
| `limit` | 1–100 | `settings.limit` |
| `hours` | 1–720 | `settings.windowHours` (0 = business-day rule) |
| `onlyActive` | bool | `settings.onlyActive` |

Out-of-range values are ignored in favour of the setting, not rejected.

```json
{
  "login": "octocat",
  "authMode": "gh-cli",
  "window": { "kind": "business-day", "label": "since last business day (Fri)",
              "hours": 72, "cutoff": "...", "now": "..." },
  "sections": [ { "id": "mine", "title": "Your open PRs", "kind": "mine",
                  "ref": "", "entries": [], "total": 4, "active": 2, "hot": 1 } ],
  "generatedAt": "...", "onlyActive": false, "limit": 25,
  "warning": "..."
}
```

Section `kind` is `mine`, `review`, `team` or `org`; `ref` carries the team slug
or org login. Entry fields are documented in
[board classification](board-classification.md).

`warning` is present on a partial GitHub failure — the board is still complete
for the tabs that resolved.

`401` without a session, `502` when GitHub failed entirely.

### `GET /api/stats`

| Parameter | Range | Default |
| --- | --- | --- |
| `days` | 1–90 | 7 |

```json
{
  "login": "octocat",
  "week": { "days": 7, "since": "...", "until": "..." },
  "opened": 6, "merged": 4, "closed": 1, "reviewed": 9,
  "reviewsWritten": 12, "approvals": 7, "repos": 3,
  "additions": 812, "deletions": 340, "filesChanged": 41,
  "kcal": 5150, "activeDays": 5, "streak": 3,
  "daily": [ { "date": "2026-08-24", "opened": 1, "merged": 0, "reviewed": 2 } ],
  "highlights": { "fastestMerge": null, "fastestMinutes": 0,
                  "biggestMerge": null, "biggestLines": 0,
                  "topRepo": "acme/api", "topRepoCount": 5 },
  "generatedAt": "...", "warning": "..."
}
```

Highlights are omitted when there is nothing to report. `reviewed` and the
`daily[].reviewed` column count different things — see
[weekly stats](weekly-stats.md#counting-reviews).

`401` without a session, `502` when GitHub failed entirely.

## Everything else

`GET /` and any unmatched non-`/api/` path serve the embedded SPA: the real file
if it exists, `index.html` otherwise. Hashed assets under `/assets/` are served
`public, max-age=31536000, immutable`; `index.html` is `no-store`.
