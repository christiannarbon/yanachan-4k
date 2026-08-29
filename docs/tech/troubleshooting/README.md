# Troubleshooting

Grouped by where the symptom shows up. Each entry is a symptom, the cause, and
the fix.

| Guide | Typical symptom |
| --- | --- |
| [Authentication](authentication.md) | the sign-in screen offers nothing, or the board is empty after signing in |
| [Networking](networking.md) | `403 unrecognised Host header`, `cross-origin request rejected`, a port already in use |
| [Build and tests](build-and-tests.md) | `make build` fails, `vue-tsc` complains about a locale key, the UI is stale after a rebuild |
| [Kubernetes](kubernetes.md) | the pod will not start, the tunnel is dead, the secret did not take |

## Before anything else

```sh
make doctor    # is the toolchain actually there
make logs      # backend and vite logs from `make dev`
```

The backend logs every `/api/` request with its duration, and prints the state
directory it resolved on startup. Two-thirds of the confusion in this project is
a state directory or a Host header, and both are in that first block of output.
