# Getting started

## What you need

| Tool | Why |
| --- | --- |
| Go 1.24+ | builds the server |
| Node 20+ and npm | builds the Vue bundle |
| `make` | every command below |
| GitHub CLI (`gh`) | optional, but the easiest way to sign in |

`make doctor` checks all of it and tells you what is missing.

## Build and run

```sh
make deps      # install Go modules and npm packages
make run       # build everything, serve on http://127.0.0.1:19080
```

Open <http://127.0.0.1:19080>. `make run` builds the frontend, embeds it into
the Go binary and runs that binary, so what you see is exactly what a production
build serves — one process, one port.

## The first screen

The dashboard asks how you want to authenticate before it does anything else.
Nothing is followed and no token is read until you say so; see
[authentication](authentication.md) for the three paths and what each one costs
you.

Once you are in, you land on [Your week](weekly-dashboard.md) at `/dashboard`.
The rest of the rail starts out nearly empty by design: teams and organizations are added by you in
[Settings](settings.md), and until then there are only your own pull requests
and the reviews requested from you.

## Working on the code

```sh
make dev       # backend on 19080, vite on 19090, both in the background
make logs      # tail both logs
make down      # stop everything
```

In dev mode, open the **vite** port — <http://localhost:19090> — not 19080. Vite
serves the frontend with hot reloading and proxies `/api` through to the backend.

```sh
make test      # go test ./... plus vue-tsc --noEmit
make check     # gofmt and go vet, then the above
make help      # every target, with a one-line description
```

## Ports

| Port | Use |
| --- | --- |
| 19080 | backend API, and the whole app in production builds |
| 19090 | vite dev server, proxies `/api` to 19080 |

Both are chosen to stay clear of the usual 8080/3000 crowd. Override them:

```sh
make run API_PORT=18080
make dev API_PORT=18080 WEB_PORT=18090
make k8s-up PF_PORT=18080     # the Kubernetes tunnel
```

## Where your data lives

Two files in a state directory — `settings.json` and `session.json`, both
`0600`. The backend prints the directory it resolved on startup. Details in
[Settings](settings.md#where-settings-are-stored).

## Next

- [Authentication](authentication.md) — sign in.
- [Your week](weekly-dashboard.md) — the tab you land on.
- [Review queues](review-queues.md) — the sections that are queues.
- [Docker](docker.md) or [Kubernetes](kubernetes.md) — run it somewhere else.
