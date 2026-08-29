# Yana-chan 4K

A pull request dashboard, and a love letter.

The dashboard: a tabbed web version of the shell script this grew out of, which
answered three questions with a `gh` search and a jq program. Same three
questions, now one tab each, plus a tab for every team and organization you
choose to follow — and a landing tab the script never had, which answers the
one question a queue cannot: what did you actually get done this week.

The love letter: this is a passion project, built as an ode to **八奈見杏菜 —
Yanami Anna**, of *負けヒロインが多すぎる！* (*Too Many Losing Heroines!*). She is
the patron saint of eating your feelings and losing with style, and she is why
this repository is named the way it is. The `4K` is a calorie count.

The default theme is hers, taken from the official site's
[calorie meter](https://makeine-anime.com/special/calorie_meter/) — the navy,
the vermillion, the yellow kcal tiles and the two typefaces are all its own.
See [Themes](#themes) and [The ode](#the-ode).

- **Your week** — the tab you land on: what you opened, merged, closed and
  reviewed over the last seven days.
- **Your open PRs** — did anyone comment inside the activity window?
- **Review requested from you** — requested from you, or already reviewed by you.
- **One tab per team** — pull requests where that team's review was requested.
- **One tab per organization** — open PRs in that org that involve you and are
  not already on another tab.

Go backend, Vue 3 frontend, one binary, two languages, ten light themes.

## Quick start

```sh
make deps      # install Go modules and npm packages
make run       # build everything, serve on http://127.0.0.1:19080
```

Open <http://127.0.0.1:19080>. The first screen asks how you want to authenticate.
Nothing is followed and no token is read until you say so.

For hot reloading while working on the code:

```sh
make dev       # backend on 19080, vite on 19090, both in the background
make logs      # tail both logs
make down      # stop everything
```

`make help` lists every target. `make doctor` checks the toolchain.

## Authentication

Three paths, all opt-in:

1. **Local gh CLI session.** On start the backend checks whether `gh` is on
   PATH and whether it is logged in — it does *not* read the token. The first
   screen reports what it found and asks for approval. Only when you approve
   does the backend run `gh auth token` and use that token for GitHub calls.
2. **OAuth device flow.** Set `GITHUB_CLIENT_ID` to an OAuth app with the device
   flow enabled. GitHub shows a code, you enter it in your browser, the backend
   polls until you approve. No client secret and no callback URL, so it behaves
   the same on a laptop, in Docker and behind a port-forward. Scopes requested:
   `repo read:org`.
3. **Token from the environment.** If `GH_TOKEN` is present the app offers it as
   an option — still behind an approval click. This is how the Docker and
   Kubernetes targets forward a local gh session into a container, since `gh`
   is not installed in the image.

The token is stored in `session.json` in the state directory with `0600`
permissions and is sent only to `api.github.com`. It is never returned to the
browser. **Sign out** in the Settings tab deletes it.

### How the server protects that token

There is no login and no cookie: the token lives on the server, so anything that
can reach the port is already authenticated. Three things keep that honest.

- **The server refuses to be framed.** Every response carries
  `X-Frame-Options: DENY` and a CSP with `frame-ancestors 'none'`. Without it a
  remote page could put the dashboard in an invisible iframe and bait a click
  onto *Use my gh CLI session* — a same-origin click, so no CSRF check would
  ever see it, and the backend would read your real token.
- **Mutating requests must be same-origin.** Anything that is not a GET or HEAD
  is rejected unless `Origin` matches, or there is no `Origin` and no browser
  fetch metadata (which is what curl looks like).
- **The `Host` header is pinned to loopback.** This is the DNS-rebinding guard.
  An attacker who points a domain of their own at `127.0.0.1` arrives with
  `Origin` and `Host` in agreement, so the origin check passes and the browser
  treats the replies as same-origin — the whole board, readable. Refusing an
  unrecognised `Host` is what stops it.

That last one is why the app answers only to `localhost`, `127.0.0.0/8` and
`::1` out of the box. To serve it under a real name, list the names:

```sh
GHDASH_ALLOWED_HOSTS=dash.example.internal make run
```

`/api/health` is exempt, so Kubernetes probes can address the pod by IP.

## Your week

The tab the app opens on. Every other tab is a queue — things waiting for you —
and this one is the opposite: a read-only look back at the last seven days.

| Figure | Where it comes from |
| --- | --- |
| Opened | `author:you created:>=` the start of the window |
| Merged | `author:you merged:>=` |
| Closed | `author:you is:unmerged closed:>=` |
| Reviewed | `reviewed-by:you -author:you`, then filtered to reviews you actually submitted inside the window |

Under those: reviews written, approvals given, repositories touched, and the
lines and files of everything merged. Then a strip per metric showing the seven
days on one shared scale, and the week's superlatives — fastest merge, biggest
merge, busiest repository, busiest day.

The window is whole local days ending with today, not a rolling 168 hours. The
chart draws a column per day and labels it with a weekday, so a window that
began at 14:07 last Sunday would put two half-Sundays at its ends and make the
first and last columns lie about themselves. `GET /api/stats?days=N` takes
anything from 1 to 90; the tab asks for 7.

Two details worth knowing before you read the numbers closely:

- **Reviewed counts pull requests, the chart counts them per day.** The tile is
  distinct branches you reviewed; a strip's column is distinct branches you
  reviewed *that day*. Go back and forth on one branch across two days and the
  strip sums to one more than the tile. Both are true, they just answer
  different questions.
- **Only submitted reviews count.** A review still open in your browser is
  `PENDING` to GitHub and is not work delivered, so it is skipped.

### The calorie figure

The hero number is a calorie total, because `4K` is a calorie count and that is
what this repository is named after. The weights live in one place,
`backend/internal/stats/types.go`:

| Event | kcal |
| --- | --- |
| Pull request opened | 200 |
| Pull request merged | 400 |
| Pull request closed unmerged | 100 |
| Review written | 150 |
| Approval given | 50 |

They are arbitrary, deliberately: shipping beats opening, reviewing somebody
else's branch counts for something, and a branch you closed unmerged still cost
you the afternoon. Under the Yanami theme the total is set in the calorie
meter's own yellow tile — see [The one exception](#the-one-exception).

### How the chart is drawn

Three strips, one per metric, over one shared scale, all in the theme's primary
(`--fact`). That is a deliberate choice rather than a stacked bar: a stack would
need three hues that stay tellable apart in all ten themes, including under
colour-vision deficiency, and splitting the series into a strip each removes the
question entirely. Each strip is a single series, so identity comes from its own
label rather than from a colour, and every bar can use the one hue the
[contrast audit](#how-a-theme-is-built) has already cleared on that surface.

Hovering a day lights its column in all three strips and writes the figures into
the chart's heading. The same numbers are in a table behind the plot, hidden
visually and read out by assistive tech.

## Teams and organizations

Both lists start empty. Add them in the Settings tab:

- Teams are written `org/team-slug`, matching the `team-review-requested:`
  search qualifier. Each one becomes a tab.
- Organizations are a bare login. Each becomes a tab of open PRs in that org
  that involve you, minus anything already shown on an earlier tab.

The Settings tab offers your actual orgs and teams as one-click suggestions,
read from your GitHub memberships.

Settings live in `settings.json` next to the session file.

## The activity window

Unchanged from the script:

| Day run | Window |
| --- | --- |
| Tue–Fri | previous 24h |
| Sat | back to Fri (24h) |
| Sun | back to Fri (48h) |
| Mon | back to Fri (72h) |

Override it with a fixed number of hours in Settings, the equivalent of the
script's `--hours`.

## Indicators

| Indicator | Meaning |
| --- | --- |
| `Reply` | somebody answered after your last comment, inside the window |
| `New` | new activity in the window, on a PR you had not commented on |
| left border, attention colour | needs your attention (a human replied, or a review is pending from you) |
| left border, primary colour | active in the window |
| left border, grey | quiet |

Those two colours follow the active [theme](#themes). Under the default Yanami
palette attention is her vermillion and active is the site's navy; under Studio
Paper they are burnt orange and teal; under a painting, whatever that painting
gave up.

Bot accounts are separated from humans using the same pattern list as the
script: `[bot]` suffixes, `dependabot`, `renovate`, `copilot`, `github-actions`,
`coderabbit`, `codecov`, `sonar`, `snyk`, `netlify`, `vercel`, `mergify`, `stale`.

### One inherited quirk

The script's review-queue branch tests `REPLY` before `NEW`, and a PR you have
never commented on has "your last comment" at time zero — so every recent
comment counts as being after it. `NEW` can therefore never fire in the review
sections; those PRs show as `Reply` instead. V1 reproduces this exactly, because
the brief was to keep the logic the same. The fix, when you want it, is one
condition in `backend/internal/board/build.go`: require `touched` before
classifying an entry as `reply`.

## Language

The whole interface is available in English and Japanese. The switch sits in the
top right beside the theme picker; the first visit follows the browser's
language preferences, and the choice is then remembered per browser under
`yana.locale`.

Copy lives in two catalogs, `frontend/src/i18n/en.ts` and `ja.ts`. English is
the source of truth: `ja.ts` is typed as its shape, so `make test` fails if a
key is added to one and not the other. Strings are read as property access
(`t.board.refresh`), not by key string, so a typo is a type error too.

Two details are worth knowing before adding copy:

- A sentence with inline markup keeps `{name}` placeholders and renders through
  `i18n/Msg.vue`, which fills each from a same-named slot. That is what lets
  Japanese put the `<code>` or `<strong>` piece where its own grammar wants it
  rather than where English left it.
- Anything with a count or a date is a function in the catalog, not a template
  with a hole in it, so each language does its own pluralising and word order.

Section titles and the window label are worded in the frontend: the backend
sends `kind` alongside its own English `label`, and a client that does not
translate can still use the label. Errors relayed from the backend or from
GitHub arrive in English in both languages.

## Themes

The picker in the top right switches the whole palette, corner radii and
typefaces. Ten themes ship, in three groups:

- **Yanami Anna — Calorie Meter**, the default, and the reason this repo exists.
- **Studio Paper**, the house palette.
- Eight derived from paintings: Cézanne, Hokusai, Hopper, Matisse, Monet, two
  Van Goghs, Wang Ximeng.

The choice is remembered per browser under `yana.art`; webfonts for a theme are
fetched only when it is first selected.

All themes are light. Components speak one semantic vocabulary -- `--panel`,
`--text`, `--fact` (primary), `--dim` (attention) -- so adding a palette is a
data change, not a refactor.

### Yanami's palette

Not invented. Every colour is one the official 負けヒロインが多すぎる！ site
declares in its own `:root`, and both faces are the pair the calorie meter page
sets its numbers in:

| Token | Where it comes from |
| --- | --- |
| `#070a7d` navy | the site's structure — rules, buttons, the 2px card outline |
| `#ff7031` vermillion | Yanami's own colour, and what the page shouts with |
| `#fff100` yellow | the kcal digit tiles |
| `#dae9f5` pale blue | the rule between rows of the calorie table |
| Quantico | the big numerals |
| Noto Sans JP | everything else |

The role assignment is the part worth explaining. Navy takes `primary`, so it
becomes `--fact` and the accent — that is what the site is built out of. The
vermillion takes `secondary`, which is what lands on `--dim`, the colour this
app paints down the left edge of a pull request that needs you. Her colour, on
the thing that is asking for attention.

### How a theme is built

```
frontend/src/styles/theme.css        what a theme does not change: spacing, motion, fallbacks
frontend/src/styles/art-themes.css   GENERATED -- one palette per theme, keyed by [data-art]
frontend/src/styles/art-themes.meta.json  names, subtitles, fonts and swatches for the picker
frontend/src/styles/art-themes.audit.md   GENERATED -- the contrast audit for every theme
frontend/src/styles/calorie-meter.css     the Yanami theme's own drawing (see below)
frontend/scripts/gen-art-themes.mjs  regenerates all three from the palettes and the token sets
```

The painting palettes come from
[art_inspired_design_system_for_AI](https://github.com/peiqingzhang/art_inspired_design_system_for_AI),
mapped onto this vocabulary and contrast-checked by the generator; Yanami and
Studio Paper are defined in the generator itself and go through the identical
pipeline, so every theme carries the same guarantees. The worst-case ratio for
every colour on every surface it is actually painted on is in
`art-themes.audit.md`. To regenerate:

```sh
cd frontend && node scripts/gen-art-themes.mjs <upstream-checkout>/themes
```

### The one exception

`calorie-meter.css` is the only file in the project where a component-level rule
knows a theme's name, and it is deliberate. What makes the calorie meter page
recognisable is not its palette but its drawing: totals set as Quantico numerals
in yellow tiles, white cards ringed in a navy outline inside the corner radius,
a vermillion cap on the figure that matters, a flat yellow marker behind the
heading. Tokens cannot express any of that.

So those four shapes live in one file, every rule behind
`[data-art="yanami-calorie-meter"]`, inert under every other theme. Nothing in
it is load-bearing: delete the file and the theme is still a complete, legible
palette — it just stops being an ode.

## Docker

```sh
make docker-build
make docker-up      # asks whether to forward your gh token, then starts compose
make docker-logs
make docker-down
```

The image is multi-stage: node builds the Vue bundle, Go embeds it into the
binary, the runtime layer is Alpine with a non-root user and a read-only root
filesystem. State lives in the `dashboard-state` volume. The port is published
on `127.0.0.1` only.

## Kubernetes

One command up, one command down.

```sh
make k8s-up      # cluster, image, secret, deploy, wait, tunnel
make k8s-down    # tunnel down, delete, wait for the namespace to go
```

`make k8s-up` is the whole sequence: it starts minikube if it is not running,
builds the image and loads it into the node, applies the secret, applies the
dev overlay, waits for the rollout, then opens a background port-forward and
polls `/api/health` until it answers. It prints the URL when it is ready.

| Target | What it does |
| --- | --- |
| `make k8s-up` | everything, then a tunnel on <http://localhost:19080> |
| `make k8s-down` | removes the workloads and the PVC, waits for the namespace |
| `make k8s-status` | pods, PVC, service, and whether the tunnel is alive |
| `make k8s-logs` | follow the pod logs |
| `make k8s-tunnel` | restart the background port-forward |
| `make k8s-untunnel` | stop it |
| `make k8s-open` | port-forward in the foreground instead |
| `make k8s-restart` | roll the pod, to pick up a changed secret or image |
| `make k8s-load` | rebuild the image and push it into the node |
| `make k8s-validate` | render both overlays and schema-check them |

The background tunnel keeps its pid in `.k8s-portforward.pid` and its output in
`.k8s-portforward.log`, both at the repo root, so `k8s-down` and `make down`
can always find and stop it.

### Layout

```
k8s/base/              namespace, PVC, deployment, ClusterIP service
k8s/overlays/dev/      image :dev, laptop-sized resources
k8s/overlays/prod/     image :v1.0.0, 1Gi storage, secret managed out of band
```

`OVERLAY` picks which one every k8s target uses, defaulting to `dev`:

```sh
make k8s-up OVERLAY=prod
```

### The token in the cluster

`make k8s-secret` runs as part of `k8s-up`. If your gh CLI is logged in it asks
before doing anything:

```
  Forward your gh CLI token into the cluster? [y/N]
```

Answer no and the app comes up with no token, ready for the OAuth device flow.
Answer yes and the token is stored in the `yana-chan-4k` secret, which the
pod reads as `GH_TOKEN` — and the app still makes you approve it on the first
screen before it is used. Set `GH_TOKEN=...` in the environment to skip the
prompt, which is also what happens when there is no terminal attached.

Changing the secret after the pod is running needs `make k8s-restart`.

The service is deliberately ClusterIP with no Ingress. The pod holds a GitHub
token and has no login of its own, so reach it through the tunnel rather than
exposing it. If you put it behind an authenticating proxy anyway, the hostname
that proxy uses has to be in `GHDASH_ALLOWED_HOSTS` or the app will refuse the
request — see [How the server protects that
token](#how-the-server-protects-that-token).

## Ports

| Port | Use |
| --- | --- |
| 19080 | backend API, and the whole app in production builds |
| 19090 | vite dev server, proxies `/api` to 19080 |

Both are chosen to stay clear of the usual 8080/3000 crowd. Override with
`make run API_PORT=...`, or `make k8s-up PF_PORT=...` for the tunnel.

## Layout

```
backend/
  cmd/server/          entry point
  internal/board/      the ported classification logic, with tests
  internal/stats/      the weekly dashboard's reduction, with tests
  internal/github/     GraphQL client and the batched PR and stats searches
  internal/ghcli/      gh CLI detection
  internal/ghauth/     OAuth device flow
  internal/api/        HTTP handlers, auth service
  internal/state/      settings and session persistence
  internal/webui/      embeds the built frontend
frontend/
  src/components/      AuthGate, TabBar, PrCard, StatsPanel, WeekChart,
                       SettingsPanel, ThemePicker, LocalePicker
  src/composables/     theme state
  src/i18n/            locale state and the English and Japanese catalogs
  src/styles/          theme tokens, the generated palettes, the calorie meter layer
  src/lib/             API client, types, time helpers
deploy/docker/         Dockerfile and compose file
k8s/base/              kustomize base
k8s/overlays/          dev and prod overlays
```

## Tests

```sh
make test    # go test ./... plus vue-tsc --noEmit
```

The board tests cover the window rule for every weekday, bot detection, and
each classification branch ported from the jq program. The stats tests cover the
whole-day window, the day buckets in a non-UTC zone, the streak rule and which
reviews count. The api tests cover the
Host allowlist, the origin guard, the security headers and the settings
validator; the state tests cover the `0600` file modes, the atomic write and
the tolerance for a corrupt session file.

## Notes

- One GraphQL request per refresh, per endpoint: every tab's search is batched
  into a single aliased query, so adding teams and orgs does not multiply round
  trips. The board and the week are two such requests, issued in parallel.
- The week uses its own, lighter fragment. The board needs comments and review
  threads to answer "who said what"; the week needs merge stamps and a diff
  size, and would otherwise pay for thirty comments per pull request that
  nothing reads.
- No LLM is called anywhere in this project.
- Light theme only, by design. No dark mode toggle.
- State lives under `yana-chan-4k` in your config directory. If a
  `github-dashboarder` directory is still there from before the rename and the
  new one is not, that one keeps being used, so upgrading does not quietly
  orphan a stored session.

## The ode

There is no practical reason a tool that answers "did anyone reply to my pull
requests" should be named after a fictional high schooler who cannot stop
eating, and there is no practical reason its default palette should be reverse
engineered out of an anime tie-in page's stylesheet. Both are true anyway.

八奈見杏菜 loses. That is the premise she is introduced with and the one she is
still living in by the end. She loses and then orders more, and the show's own
site keeps a straight-faced running tally of it, episode by episode, scene by
scene, in kcal and carbohydrates, as though the committee commissioned a
nutritional survey. Somebody built that page. Somebody picked Quantico for the
digits and gave every one of them its own yellow tile.

This is for her. Everything else here — the batched GraphQL, the contrast
generator, the Kubernetes overlays — is scaffolding around the part where the
tab counter is a calorie counter.

負けヒロインが多すぎる！ is © Takibi Amamori / Shogakukan / "Makeine" Production
Committee. Nothing here is affiliated with or endorsed by them; the colours and
type choices are cited the way you would cite a source, and the artwork is not
redistributed. Go watch it.
