# Frontend

Vue 3 with `<script setup>`, TypeScript, Vite. No router, no state library, no
component framework — the whole app is one page with a tab bar, and the state
that outlives a component fits in two modules.

## Layout

```
frontend/
  src/components/      AuthGate, TabBar, PrCard, StatsPanel, WeekChart,
                       SettingsPanel, ThemePicker, LocalePicker, CornerControls
  src/composables/     useTheme
  src/i18n/            useI18n, Msg.vue, and the en / ja catalogs
  src/styles/          theme tokens, the generated palettes, the calorie meter layer
  src/lib/             API client, types, time helpers
  scripts/             the theme generator and its colour maths
```

## App.vue owns the data

`App.vue` is the only component that talks to the API. It holds `authStatus`,
`board`, `stats` and `settings` as refs, and passes them down. Components emit
back up; nothing fetches on its own.

Bootstrap: ask `/api/auth/status`; if unauthenticated, render `AuthGate` and
stop. If authenticated, load settings and refresh.

A refresh fires both endpoints together — `/api/board` and `/api/stats` — with
their errors kept in **separate** refs. A failing week must not blank the
board's own notice, and the dashboard tab reports its own trouble in place.

Optional auto-refresh runs every 5 minutes; a separate 30-second timer advances
the clock that "3 hours ago" is rendered against, so relative stamps do not go
stale on a tab left open.

`FIXED_TABS` (`dashboard`, `settings`) exist so a reload never navigates away
from a tab when the section list changes underneath it.

## The two pieces of global state

Both are module-level refs with a `localStorage` mirror, both readable from any
component, neither of them a store library.

| Module | Holds | Key |
| --- | --- | --- |
| `composables/useTheme.ts` | the active theme, written to `[data-art]` on the root | `yana.art` |
| `i18n/index.ts` | the active locale and the catalog it selects | `yana.locale` |

Both read a pre-rename key as a fallback and both wrap every `localStorage`
access in `try/catch` — private windows and blocked site data throw rather than
returning null.

Theme webfonts are loaded lazily, on first selection, so nine of the ten
palettes cost nothing until someone tries them.

## Components

| Component | Responsibility |
| --- | --- |
| `AuthGate` | the first screen: what sign-in paths exist, and the device-flow polling |
| `TabBar` | the tab strip, including the per-team and per-org tabs |
| `StatsPanel` | the week's tiles, the calorie figure, the superlatives |
| `WeekChart` | three strips over one shared scale, plus the visually-hidden table |
| `PrCard` | one pull request: indicators, the left border, the actors line |
| `SettingsPanel` | teams, orgs, view settings, session |
| `ThemePicker` / `LocalePicker` | the two pickers |
| `CornerControls` | the corner cluster the two pickers live in, shared with `AuthGate` |

## Where styling stops being a component's business

Components speak one semantic vocabulary — `--panel`, `--text`, `--fact`
(primary), `--dim` (attention) — and never a colour literal or a theme name.
That is what makes adding a palette a data change.

There is exactly one deliberate exception, `styles/calorie-meter.css`, and it is
documented in [theming system](../low-level-design/theming-system.md#the-one-exception).

## Types

`src/lib/types.ts` mirrors the backend's JSON by hand. There is no code
generation: the surface is a dozen structs, and the drift is caught by
`vue-tsc` the moment a field is read that the type does not have.

```sh
make test    # includes vue-tsc --noEmit
```

## Dev server

Vite on 19090, proxying `/api` to 19080 with `changeOrigin: false` — the
browser's `Host` header survives the proxy, which is what lets the backend's
[host allowlist](security.md) see a real hostname in development. The backend is
told about the dev origin through `GHDASH_DEV_ORIGIN`, which `make dev` sets.
