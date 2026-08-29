# Configuration reference

There are no flags and no config file. Every knob is an environment variable, so
Docker and Kubernetes need nothing the binary does not already understand.

`backend/internal/config/config.go`

## Server

| Variable | Default | Effect |
| --- | --- | --- |
| `GHDASH_ADDR` | `127.0.0.1:19080` | listen address. Binding beyond loopback needs `GHDASH_ALLOWED_HOSTS` set too |
| `GHDASH_ALLOWED_HOSTS` | *(empty)* | comma-separated extra `Host` values, beyond loopback. See [the security model](../architecture/security.md#guard-3--the-host-header-is-pinned-to-loopback) |
| `GHDASH_DEV_ORIGIN` | *(empty)* | one permitted CORS origin, for the vite dev server. `make dev` sets it; never set it in a deployment |
| `GHDASH_STATE_DIR` | OS config dir + `yana-chan-4k` | where `settings.json` and `session.json` live |

## GitHub

| Variable | Default | Effect |
| --- | --- | --- |
| `GITHUB_CLIENT_ID` | *(empty)* | enables the OAuth device flow. Unset means the option is not offered, and the server says so on startup |
| `GHDASH_ALLOW_GH_CLI` | `true` | whether the local `gh` session may be offered at all. `false` in the image and the pod |
| `GHDASH_GITHUB_TOKEN` | *(empty)* | a token from the environment — checked first |
| `GH_TOKEN` | *(empty)* | …then this. What the Docker and k8s targets forward |
| `GITHUB_TOKEN` | *(empty)* | …then this |
| `GHDASH_GRAPHQL_ENDPOINT` | `https://api.github.com/graphql` | override for testing against a fake |
| `GHDASH_GITHUB_WEB` | `https://github.com` | device-flow endpoints are derived from this |

A token in the environment is still only *offered*. The user approves it on the
first screen before it is used.

## Defaults

| Variable | Default | Effect |
| --- | --- | --- |
| `GHDASH_LIMIT` | `25` | starting value for the per-query pull request cap. Once settings are saved, `settings.json` wins |

## Parsing

- Empty and whitespace-only values are treated as unset, so `GH_TOKEN=` in a
  compose file does not count as a token.
- `GHDASH_ALLOW_GH_CLI` is `strconv.ParseBool` — `1`, `t`, `T`, `true`, `TRUE`
  and their false counterparts. An unparseable value falls back to the default
  rather than failing.
- `GHDASH_LIMIT` must be a positive integer; anything else falls back.
- `GHDASH_ALLOWED_HOSTS` is split on commas with empty entries dropped. Port
  suffixes are stripped when the allowlist is built, so
  `dash.example.internal:19080` and `dash.example.internal` are the same entry.

## Make variables

Not environment variables for the binary — arguments to the Makefile.

| Variable | Default | Used by |
| --- | --- | --- |
| `API_PORT` | `19080` | `run`, `dev`, `docker-up` |
| `WEB_PORT` | `19090` | `dev` (vite) |
| `API_HOST` | `127.0.0.1` | `run`, `dev` |
| `IMAGE_NAME` / `IMAGE_TAG` | `yana-chan-4k` / `dev` | `images`, `docker-*`, `k8s-load` |
| `OVERLAY` | `dev` | every `k8s-*` target |
| `PF_PORT` | `$(API_PORT)` | the Kubernetes port-forward |

```sh
make run API_PORT=18080
make k8s-up OVERLAY=prod PF_PORT=18080
```

## Where each deployment sets what

| | `make run` | Docker | Kubernetes |
| --- | --- | --- | --- |
| `GHDASH_ADDR` | `127.0.0.1:19080` | `:19080` | `:19080` |
| `GHDASH_STATE_DIR` | OS config dir | `/data` (volume) | `/data` (PVC) |
| `GHDASH_ALLOW_GH_CLI` | `true` | `false` | `false` |
| `GH_TOKEN` | inherited | forwarded on approval | from the `yana-chan-4k` secret, optional |
| `GITHUB_CLIENT_ID` | inherited | passed through compose | from the secret, optional |
