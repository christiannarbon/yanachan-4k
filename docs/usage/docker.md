# Docker

```sh
make docker-build
make docker-up      # asks whether to forward your gh token, then starts compose
make docker-logs
make docker-down
```

`make docker-up` publishes the dashboard on <http://127.0.0.1:19080>. Override
with `make docker-up API_PORT=18080`.

## The image

Multi-stage, and small at the end of it:

1. **node** builds the Vue bundle.
2. **golang** embeds that bundle into the server binary with `go:embed`, built
   `CGO_ENABLED=0` and trimmed.
3. **alpine** carries the binary, `ca-certificates` and `tzdata` — no node, no
   web server, nothing to serve the frontend with because the frontend is inside
   the binary.

The runtime layer runs as a non-root user (uid 10001) and the port is published
on `127.0.0.1` only. Base images are pinned by digest rather than tag, so a
rebuilt upstream cannot silently change what goes into the binary. Refresh a
digest with:

```sh
docker buildx imagetools inspect node:24-alpine --format '{{.Manifest.Digest}}'
```

There is a `HEALTHCHECK` hitting `/api/health` every 30 seconds.

## The token

`gh` is not installed in the image, so the container cannot inherit your local
session the way `make run` does. `make docker-up` offers to forward it instead:

```
  Forward your gh CLI token into the container? [y/N]
```

Answer yes and the token goes in as `GH_TOKEN` — and the app still makes you
approve it on the first screen before it is used. Answer no and the container
comes up with no token, ready for the [OAuth device flow](authentication.md#2-oauth-device-flow):

```sh
GITHUB_CLIENT_ID=Iv1.xxxx make docker-up
```

`GHDASH_ALLOW_GH_CLI` is `false` in the image, so the gh CLI option is never
offered inside a container even if something mounts a `gh` binary in.

## State

State lives on the `dashboard-state` volume, mounted at `/data`. It survives
`make docker-down`; it does not survive `docker compose down -v`.

## Compose file

`deploy/docker/docker-compose.yml`. The build context is the repository root, so
run compose through the Makefile targets or from `deploy/docker/` with
`GHDASH_IMAGE` set.
