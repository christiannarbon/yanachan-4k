# Review queues

Every section other than [Your week](weekly-dashboard.md) is a queue: pull
requests that are waiting on somebody, sorted so the ones waiting on *you* are
at the top.

| Section | What is on it |
| --- | --- |
| **Your open PRs** | your branches — did anyone comment inside the activity window? |
| **Review requested from you** | requested from you, or already reviewed by you |
| **One per team** | pull requests where that team's review was requested |
| **One per organization** | open PRs in that org that involve you and are not already in another section |

Team and organization sections appear only once you add them in
[Settings](settings.md).

## Finding a queue

The sections live in the rail down the left of the page, under three headings:
your queues, then your teams, then your organizations. Click a heading to fold
its list away — handy once you follow a dozen orgs — and the fold is remembered
per browser. On a narrow window the rail becomes a drawer behind the
**Sections** button in the header.

Every section has its own address, so a queue can be bookmarked or sent to
somebody, and back and forward work:

| Path | Shows |
| --- | --- |
| `/dashboard` | [Your week](weekly-dashboard.md) |
| `/prs/mine` | your open pull requests |
| `/prs/review` | reviews requested from you |
| `/prs/team/acme/platform` | a followed team |
| `/prs/org/acme` | a followed organization |
| `/settings` | [Settings](settings.md) |

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

## Grouped by repository

A queue is drawn as one group per repository rather than one long column. Each
heading carries the repository's name, how many of its pull requests are in the
section, and the attention dot if any of them need you.

Click a heading to fold that repository away. Folds are remembered per browser
and shared across the queues, so a repository you would rather not look at stays
folded everywhere and after a reload. **Collapse all**, beside the section's
own heading, folds every group at once — and becomes **Expand all** once they
all are.

On a wide window each group's cards wrap into columns rather than stretching to
the far edge; a laptop keeps the single column it had.

Grouping rearranges the list without re-ranking it. A group takes the position
of its first pull request and the order inside a group is untouched, so the
sort still holds: the repository you are needed in leads the page.

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
