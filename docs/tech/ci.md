# Continuous integration

`.github/workflows/ci.yml`

Runs on every pull request to `main`, every push to `main`, and on demand
(`workflow_dispatch`). Permissions are `contents: read` — nothing in the
pipeline writes to the repository, and the image job builds without pushing.

A new push to a pull request cancels the previous run; runs on `main` are left
alone so the record of what passed stays intact.

## The jobs

| Job | What it proves |
| --- | --- |
| **Backend** | `gofmt` clean, `go vet` clean, `go test -race` green |
| **Frontend** | the lockfile installs, `vue-tsc` passes, the bundle builds |
| **Vulnerabilities** | `govulncheck` finds no reachable standard library vulnerability |
| **Build and smoke test** | `make build` produces a binary that serves the embedded app correctly |
| **Kubernetes manifests** | both overlays render, schema-check, and keep their security posture |
| **Container image** | the multi-stage build works and the container behaves |
| **Documentation** | every internal link in the Markdown resolves |

They run in parallel. Nothing depends on anything else.

## Why each one is there

**Backend.** `-race` because `state.Store` is the only shared mutable state in
the process, guarded by an `RWMutex`, and a lost lock there is exactly the bug a
serial test run passes.

**Frontend.** `npm ci` rather than `npm install`, so a `package.json` that has
drifted from the lockfile fails here instead of in the image build. The type
check runs as its own step even though `npm run build` type-checks first,
because it is where a locale key added to `en.ts` and not `ja.ts` fails, and
that deserves to report as a type error. See [i18n](low-level-design/i18n.md).

**Vulnerabilities.** The module has no third-party dependencies, so this is a
standard library scan and the toolchain pin in `go.mod` is the whole of what it
scans. Not a small surface here: the one place this app talks to the network
holding a GitHub token goes through `crypto/tls`, `crypto/x509` and `net/http`.
`govulncheck` reports only what the code actually reaches, which is what makes
it worth gating on.

**Build and smoke test.** The two halves compile in separate jobs; neither
proves the thing this project ships, which is one binary with the Vue bundle
inside it. So this job runs `make build`, starts the binary, and asserts over
HTTP that health answers, the frontend is embedded, an unmatched API path
answers a JSON 404, `/api/board` refuses an unauthenticated caller, the
[Host allowlist](architecture/security.md#guard-3--the-host-header-is-pinned-to-loopback)
refuses an unknown name, and health is exempt from that allowlist.

The embed assertion is not redundant with a successful compile: only `.gitkeep`
is tracked in `backend/internal/webui/dist`, so a binary built without a staged
bundle compiles fine and then answers `frontend bundle is missing`.

The last two assertions cover the DNS-rebinding guard end to end. It has unit
tests at the handler level; this is the one place it is exercised in the built
binary.

**Kubernetes manifests.** `kustomize` and `kubeconform` are downloaded at pinned
versions rather than taken from the runner image, because what is preinstalled
there changes without notice. Both overlays are rendered — `prod` is the one
nobody renders by hand, so it is the one that quietly breaks.

Beyond the schema, two assertions about things that fail silently: the
[pod hardening](../usage/kubernetes.md#pod-hardening) survives, and neither
overlay grows an `Ingress`, a `LoadBalancer` or a `NodePort`. The pod holds a
GitHub token and has no login of its own, so ClusterIP is a decision rather than
a default.

**Container image.** The only job that exercises the digest-pinned base images.
Then it checks the container runs as uid 10001 — a rebuild that drops the `USER`
line passes every other check in the file — and that `GHDASH_ALLOW_GH_CLI` is
still `false` in the image environment.

**Documentation.** `make docs-check`, the same command you can run locally.

## Running the same checks locally

```sh
make check         # gofmt, go vet, go test, vue-tsc, and the doc links
make docs-check    # just the documentation links
make build         # what the build job builds
make k8s-validate  # renders and schema-checks both overlays
make docker-build  # the image
```

`make check` is the closest single command to what CI runs. It does not include
`govulncheck` or the smoke tests; those are:

```sh
cd backend && go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
```

## Pinned versions

Everything the pipeline installs is pinned, so a green tick means the same check
it meant last week.

| Tool | Version | Where |
| --- | --- | --- |
| Go | from `backend/go.mod` | the `toolchain` line is the single pin |
| Node | 24 | matching the Dockerfile's builder |
| govulncheck | `v1.1.4` | workflow |
| kustomize | `v5.4.3` | workflow |
| kubeconform | `v0.6.7` | workflow, `-kubernetes-version 1.33.0` |

[Dependabot](../../.github/dependabot.yml) moves the action versions weekly, the
Dockerfile's base image digests weekly, and the frontend packages monthly as one
grouped pull request. There is no `gomod` entry: the module has no third-party
dependencies.

## Artefacts

The build job uploads the Linux binary as `yana-chan-4k-linux-amd64`, kept for
seven days. Useful for reproducing a smoke-test failure without building
locally.

## Troubleshooting a red run

Most failures map onto an existing guide, since every job runs a command you can
run yourself:

| Job | Start at |
| --- | --- |
| Backend, Frontend, Build | [build and tests](troubleshooting/build-and-tests.md) |
| Kubernetes manifests | [Kubernetes](troubleshooting/kubernetes.md) |
| Container image | [Docker](../usage/docker.md) |
| Documentation | the checker names the file and the link |

`govulncheck` failing usually means the toolchain pin has aged. Its output names
the fixed-in version; bump `toolchain` in `backend/go.mod` and check the
Dockerfile's builder digest is at least that version.
