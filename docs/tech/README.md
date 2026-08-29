# Tech documentation

Yana-chan 4K is a Go backend and a Vue 3 frontend that ship as one binary: the
built frontend is embedded into the server with `go:embed`, so production is a
single process listening on one port.

## [Architecture](architecture/)

The shape of the system and the reasoning behind it.

| Document | What it answers |
| --- | --- |
| [Overview](architecture/overview.md) | what the pieces are and how a refresh flows through them |
| [Backend](architecture/backend.md) | what each Go package owns, and what it deliberately does not |
| [Frontend](architecture/frontend.md) | components, state, and where the styling boundary sits |
| [Security model](architecture/security.md) | the token has no login in front of it — what keeps that safe |

## [Low level design](low-level-design/)

The rules inside a package, at the level you need before editing one.

| Document | What it answers |
| --- | --- |
| [API reference](low-level-design/api-reference.md) | every endpoint, its shape, and its failure modes |
| [Activity window](low-level-design/activity-window.md) | the business-day rule and the fixed override |
| [Board classification](low-level-design/board-classification.md) | how a pull request becomes `Reply`, `New` or quiet |
| [Weekly stats](low-level-design/weekly-stats.md) | the reduction behind the landing tab, and the kcal weights |
| [GraphQL batching](low-level-design/graphql-batching.md) | one request per refresh, however many tabs |
| [State persistence](low-level-design/state-persistence.md) | `settings.json`, `session.json`, file modes, atomic writes |
| [Theming system](low-level-design/theming-system.md) | tokens, the generator, the contrast audit, the one exception |
| [Internationalization](low-level-design/i18n.md) | two catalogs, checked by the type system |
| [Configuration reference](low-level-design/configuration-reference.md) | every environment variable |

## [Troubleshooting](troubleshooting/)

Symptom-first, grouped by where the symptom appears: [authentication](troubleshooting/authentication.md),
[networking](troubleshooting/networking.md), [build and tests](troubleshooting/build-and-tests.md),
[Kubernetes](troubleshooting/kubernetes.md).

## Working on the code

```sh
make deps      # go mod download + npm install
make dev       # backend on 19080, vite on 19090, both in the background
make logs      # tail both
make check     # gofmt, go vet, go test ./..., vue-tsc --noEmit
make down      # stop everything
```

Two rules hold across the codebase and are worth knowing before the first edit:

- **The token never reaches the browser.** Anything you add to a response is
  something the browser can read; `backend/internal/api/server_test.go` asserts
  the token is not in `/api/auth/status`.
- **Themes are data.** Components speak `--panel`, `--text`, `--fact`, `--dim`.
  A component that names a theme is a bug, with exactly one deliberate
  exception — see [theming system](low-level-design/theming-system.md).
