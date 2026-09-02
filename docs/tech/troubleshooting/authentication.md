# Troubleshooting: authentication

## The first screen offers nothing

All three paths are opt-in and each has a precondition.

| Path | Precondition | Check |
| --- | --- | --- |
| gh CLI | `gh` on `PATH`, logged in, and `GHDASH_ALLOW_GH_CLI` not `false` | `gh auth status` |
| OAuth device flow | `GITHUB_CLIENT_ID` set | the startup log says if it is not |
| Environment token | `GHDASH_GITHUB_TOKEN`, `GH_TOKEN` or `GITHUB_TOKEN` non-empty | `echo $GH_TOKEN` |

The backend prints this on startup when the second is missing:

```
GITHUB_CLIENT_ID is unset: OAuth device sign-in is disabled
```

Inside a container all three can be false at once: the image sets
`GHDASH_ALLOW_GH_CLI=false` and forwards a token only if you agreed to it. Run
`make docker-up` again and answer `y`, or pass `GITHUB_CLIENT_ID`.

## "gh CLI is installed but not logged in"

```sh
gh auth login
```

Then reload the page. The status is re-read per request, so no restart is needed.

## "reading the local gh CLI session is disabled by configuration"

`GHDASH_ALLOW_GH_CLI=false`. That is the default in the image and the pod, and
it is deliberate — a container should not inherit an ambient session. Use the
device flow or an environment token there.

## "no token was supplied by the environment"

The variable is empty or whitespace. Note that `GH_TOKEN=` in a compose file
counts as unset: empty values are treated as absent.

Under Kubernetes, check the secret actually has content:

```sh
kubectl -n yana-chan-4k get secret yana-chan-4k -o jsonpath='{.data.gh-token}' | base64 -d | wc -c
```

Zero means `make k8s-secret` ran without a token. Re-run it, then
`make k8s-restart` — the pod reads the secret at startup.

## "token check failed"

The token was read but GitHub rejected it or the call did not complete.

- Expired or revoked token — `gh auth refresh`, or issue a new one.
- No network route to `api.github.com` from wherever the server is running.
- GitHub Enterprise: `GHDASH_GRAPHQL_ENDPOINT` and `GHDASH_GITHUB_WEB` both need
  pointing at your instance.

## The device flow never completes

`POST /api/auth/device/poll` returns a `state`:

| `state` | Meaning |
| --- | --- |
| `pending` | you have not approved it in the browser yet |
| `slow_down` | you polled too fast; the UI backs off automatically |
| `expired` | the code timed out — start again |
| `denied` | you declined it on github.com |

"no device authorization is in progress" means the server has no pending flow —
the backend restarted, or the poll arrived before a start. Click sign-in again.

## Signed in, but the board is empty

- Check whether the sections exist at all. Team and org sections only appear
  once you add them in [Settings](../../usage/settings.md).
- `Your open PRs` genuinely empty is a correct answer if you have no open pull
  requests.
- A `warning` on the board means a partial GitHub failure — one search resolved,
  another did not. The message is GitHub's own.

## Suggestions do not load in Settings

```
Could not list your memberships: ...
```

Almost always a token without the `read:org` scope. Fine-grained tokens need an
organization read permission. You can still type teams and orgs in by hand;
suggestions are a convenience.

## Signing out did not sign me out

`POST /api/auth/logout` deletes `session.json`. If you are immediately signed in
again, you are re-approving an *environment* token that is still present —
`GH_TOKEN` in your shell, the compose file, or the secret. Remove it there.
