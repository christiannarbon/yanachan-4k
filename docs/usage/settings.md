# Settings

The Settings tab. Everything here is stored on the server, in the state
directory, and applies to whoever can reach the port.

## Teams to follow

Written `org/team-slug`, matching GitHub's `team-review-requested:` search
qualifier. Each one becomes a tab listing open pull requests where a review is
requested from that team.

The panel offers your actual teams as one-click suggestions, read from your
GitHub memberships. If that list fails to load — a token without `read:org`,
usually — the warning is shown and you can still type a team in by hand.

Up to 200 teams. A slash is required; `my-team` on its own is rejected.

## Organizations to follow

A bare login, no slash. Each becomes a tab of open PRs in that org that involve
you, minus anything already shown on an earlier tab, so the same pull request
never appears twice.

Also offered as suggestions, also capped at 200.

## View

| Control | Effect |
| --- | --- |
| **Pull requests per query** | 1 to 100, the equivalent of the script's `--limit`. This is per queue tab, not in total, and it does not touch Your Week. |
| **Activity window** | Hours, 0 to 720. Leave it at 0 for the [business-day rule](review-queues.md#the-activity-window). |
| **Hide pull requests with no new activity** | drops the quiet ones from every queue |
| **Show the full URL under each pull request** | on by default |

Anything out of range is corrected on save rather than rejected: a limit outside
1–100 falls back to the configured default, a window outside 0–720 falls back to
the business-day rule.

## Session

Shows who you are signed in as and by which of the three
[authentication](authentication.md) paths.

**Sign out and forget the token** deletes `session.json`. Nothing else is
touched — your teams, orgs and view settings survive.

## Where settings are stored

Two files, both `0600`, in a state directory the backend prints on startup:

| File | Holds |
| --- | --- |
| `settings.json` | teams, orgs, and everything under **View** |
| `session.json` | how you authenticated and the token |

The directory is `yana-chan-4k` inside your OS config directory:

| Platform | Path |
| --- | --- |
| macOS | `~/Library/Application Support/yana-chan-4k/` |
| Linux | `~/.config/yana-chan-4k/` |
| Windows | `%AppData%\yana-chan-4k\` |

Override it with `GHDASH_STATE_DIR`.

The app was called `github-dashboarder` before it was called Yana-chan 4K. If
that older directory is still there and the new one is not, the old one keeps
being used — upgrading does not quietly orphan a stored session. Once the new
directory exists it always wins.

Under Docker, state lives on the `dashboard-state` volume; under Kubernetes, on
the PVC — and `make k8s-down` deletes that PVC, session included.
