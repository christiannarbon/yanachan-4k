# Troubleshooting: networking

Most of what goes wrong here is one of the three guards in
[the security model](../architecture/security.md) doing its job.

## `403 unrecognised Host header`

```
unrecognised Host header; set GHDASH_ALLOWED_HOSTS to serve this app under that name
```

The app answers only to `localhost`, `127.0.0.0/8` and `::1`, plus the address
it was told to listen on, plus anything you list. You reached it under some
other name — a machine hostname, a reverse proxy, a `/etc/hosts` alias.

```sh
GHDASH_ALLOWED_HOSTS=dash.example.internal make run
```

Comma-separate several. Port suffixes are stripped, so listing the bare hostname
is enough.

This is the DNS-rebinding guard, and it is the one you are most likely to hit
legitimately and most likely to regret disabling. Read
[guard 3](../architecture/security.md#guard-3--the-host-header-is-pinned-to-loopback)
before you widen it — and if you widen it for a proxy, that proxy must
authenticate.

`/api/health` is exempt, so a probe answering while the UI 403s is expected, not
a contradiction.

## `403 cross-origin request rejected`

A `PUT` or `POST` arrived with an `Origin` that is not ours. Causes, in order of
likelihood:

1. **Dev server without `GHDASH_DEV_ORIGIN`.** Vite is on 19090 and the backend
   on 19080, so saving settings is cross-origin. `make dev` sets the variable;
   starting the backend by hand does not. Use `make dev`, or set it:

   ```sh
   GHDASH_DEV_ORIGIN=http://localhost:19090 go run ./cmd/server
   ```

2. **A proxy rewriting `Origin`.** The vite proxy is configured
   `changeOrigin: false` on purpose. A proxy that rewrites the header will break
   this guard.

3. **A browser extension** injecting requests.

`curl` is *allowed* to mutate — no `Origin` and no fetch metadata is what a
command-line client looks like, and keeping the API scriptable from the same
machine is deliberate. So `curl -X PUT` succeeding while the browser is refused
is expected.

## Blank page, or the browser will not run the app

Check the console for a CSP violation. The policy is strict and set in
`api.setSecurityHeaders`. If you added a CDN font, an external script or an
inline handler, the policy is what is refusing it — add it there rather than
loosening the whole policy.

`frame-ancestors 'none'` and `X-Frame-Options: DENY` mean the app cannot be
embedded in an iframe at all. That is intentional and is the single most
important header in the project.

## `404 no such endpoint: GET /api/...`

The route does not exist. Note that this arrives as **JSON**, deliberately —
without the catch-all, an unmatched API path would fall through to the SPA
handler and return `index.html` with a 200, and the frontend would try to
`JSON.parse` a web page.

Check the method too: routes are registered as `GET /api/board`, so a `POST`
there is a 405.

## `404 frontend bundle is missing; run 'make build-frontend'`

The Go binary was built without a staged frontend. `make run` does both; `go run
./cmd/server` on a clean checkout serves the committed placeholder.

```sh
make build-frontend
```

## Port already in use

```sh
make down            # stops dev servers, compose and the k8s tunnel
make run API_PORT=18080
```

`make down` is the blanket stop, and it also kills the background port-forward
using `.k8s-portforward.pid` at the repository root.

## The UI is stale after a rebuild

`index.html` is served `no-store` and hashed assets `immutable`, so a hard
reload should never be necessary. If it is, you are probably looking at a
`make run` binary built before your last frontend change — rebuild.
