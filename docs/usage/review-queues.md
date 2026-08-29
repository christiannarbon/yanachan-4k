# Review queues

Every tab other than [Your week](weekly-dashboard.md) is a queue: pull requests
that are waiting on somebody, sorted so the ones waiting on *you* are at the top.

| Tab | What is on it |
| --- | --- |
| **Your open PRs** | your branches — did anyone comment inside the activity window? |
| **Review requested from you** | requested from you, or already reviewed by you |
| **One tab per team** | pull requests where that team's review was requested |
| **One tab per organization** | open PRs in that org that involve you and are not already on another tab |

Team and organization tabs appear only once you add them in
[Settings](settings.md).

## The activity window

The window decides what counts as recent. Unchanged from the shell script this
grew out of:

| Day run | Window |
| --- | --- |
| Tue–Fri | previous 24h |
| Sat | back to Fri (24h) |
| Sun | back to Fri (48h) |
| Mon | back to Fri (72h) |

The point is that a Monday morning should show you the weekend's worth of
activity rather than the previous 24 hours of nothing.

Override it with a fixed number of hours in Settings — the equivalent of the
script's `--hours`. The exact arithmetic is in
[activity window](../tech/low-level-design/activity-window.md).

## Indicators

| Indicator | Meaning |
| --- | --- |
| `Reply` | somebody answered after your last comment, inside the window |
| `New` | new activity in the window, on a PR you had not commented on |
| left border, attention colour | needs your attention (a human replied, or a review is pending from you) |
| left border, primary colour | active in the window |
| left border, grey | quiet |

Those two colours follow the active [theme](themes.md). Under the default Yanami
palette attention is her vermillion and active is the site's navy; under Studio
Paper they are burnt orange and teal; under a painting, whatever that painting
gave up.

## Bots

Bot accounts are separated from humans so a `dependabot` comment never makes a
pull request look like it needs you. The pattern list is the script's:

`[bot]` suffixes, `dependabot`, `renovate`, `copilot`, `github-actions`,
`coderabbit`, `codecov`, `sonar`, `snyk`, `netlify`, `vercel`, `mergify`,
`stale`.

## One inherited quirk

`New` never fires in the review sections. Pull requests that would qualify show
as `Reply` instead.

The script's review-queue branch tests `REPLY` before `NEW`, and a PR you have
never commented on has "your last comment" at time zero — so every recent
comment counts as being after it. V1 reproduces this exactly, because the brief
was to keep the logic the same.

The fix, when you want it, is one condition in `backend/internal/board/build.go`.
See [board classification](../tech/low-level-design/board-classification.md#the-inherited-quirk).

## Only active

Settings has an **only active** toggle, which drops the quiet pull requests and
leaves the ones that moved inside the window.
