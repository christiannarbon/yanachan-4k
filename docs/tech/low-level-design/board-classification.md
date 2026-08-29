# Board classification

`backend/internal/board/build.go`

A direct port of the jq program from the shell script this dashboard grew out
of: same event sources, same bot detection, same window, same `REPLY`/`NEW`
semantics, same sort order. Where the port and the script disagree, the port is
wrong — including in the one place it is
[deliberately faithful to a bug](#the-inherited-quirk).

## The event timeline

Three sources are flattened into one list, exactly like the jq `evs` function:

| Source | GraphQL field | Kind |
| --- | --- | --- |
| Issue comments | `comments(last:30)` | `comment` |
| Reviews | `reviews(last:30)` | `review` (carries `state`) |
| Review thread comments | `reviewThreads(last:20).comments(first:20)` | `thread` |

```go
type event struct {
    who   string   // login, or "ghost" for a deleted account
    bot   bool
    at    time.Time
    kind  string
    state string   // reviews only
}
```

The `last:` bounds are the script's. A pull request with four hundred comments
is classified on its most recent thirty, which is the right answer for "did
somebody reply to me" and the wrong answer for an archaeology tool. This is not
an archaeology tool.

## Bot detection

`bots.go`, one regexp, case-insensitive:

```
\[bot\]|-bot$|^bot$|^dependabot|^renovate|^copilot|^github-actions|
^coderabbit|^codecov|^sonar|^snyk|^netlify|^vercel|^mergify|^stale
```

Plus anything GitHub itself types as `Bot`. Bots are counted and named
separately from humans throughout, and only humans make an entry `Hot`.

## Two classifiers

### `mineEntry` — your own pull requests

The question is *did anybody else touch this inside the window*.

```
others  = events not by you
recent  = others after cutoff
humans  = recent, not bots

Active = len(recent) > 0
Hot    = len(humans) > 0
Status = new    if Active
         quiet  otherwise
```

`LastActivityAt` is the newest of `recent` when active, and the newest of
`others` when quiet — so a quiet pull request still shows when it last moved.

### `reviewEntry` — somebody else's, in your queue

The question is *is this waiting on me*.

```
mine    = events by you
others  = events not by you
myLast  = max(mine.at)            zero time if you never commented
touched = len(mine) > 0

replies = others after myLast AND after cutoff
recent  = others after cutoff

Active = len(replies) > 0 || (!touched && (len(recent) > 0 || pendingMe || pendingTeam))
Hot    = len(replyHumans) > 0 || (!touched && pendingMe)
```

`pendingMe` and `pendingTeam` come from `reviewRequests`: a review requested
from you personally, or from the team whose tab this is.

Status, in order:

| Branch | Status |
| --- | --- |
| `len(replies) > 0` | `reply` |
| `!touched && len(recent) > 0` | `new` |
| otherwise | `quiet` |

Then the review-queue extras: `YourState` (`approved`, `changes_requested`, or
`commented` for anything else) and `YourLastAt` when you have touched it;
`Awaiting` (`you` or `team`) when you have not; `AlsoRequestedFromYou` when you
are on a team tab *and* personally requested.

## The inherited quirk

**`new` can never fire in the review sections.** Pull requests that would
qualify show as `reply` instead.

The reason is the branch order above combined with `myLast` being the zero time
when you have never commented. Every event in the universe is after the zero
time, so `replies` — which is `others after myLast and after cutoff` — collapses
to `recent`, and the first branch always wins.

The script does this. V1 reproduces it exactly, because the brief was to keep
the logic the same and a dashboard that quietly disagrees with the script it
replaced is worse than one that is faithfully odd.

The fix, when you want it, is one condition:

```go
case touched && len(replies) > 0:      // require touched
    e.Status = StatusReply
```

`TestReviewEntryDetectsReplyAfterYourComment` covers the current behaviour, so
that change will fail a test — which is the point. Change the test in the same commit and say why.

## Sections

Six kinds of tab, assembled in this order, because the order is what prevents
duplicates:

1. **`mine`** — `is:open is:pr archived:false author:me sort:updated-desc`.
2. **Team tabs** — one per followed team, `team-review-requested:org/slug`. Built
   *before* the personal review queue: a team-requested PR belongs to the team
   tab and is dropped from the personal one, which is what the script does.
3. **`review`** — the union of `user-review-requested:me` and
   `reviewed-by:me -author:me`, minus anything already on a team tab, deduped by
   URL.
4. **Org tabs** — one per followed org, `org:X involves:me`, minus everything on
   every earlier tab. Within an org tab, your own pull requests go through
   `mineEntry` and everyone else's through `reviewEntry`.

Every search carries the base `is:open is:pr archived:false` and
`sort:updated-desc`.

## Sorting and counts

```
hot (0) before active (1) before quiet (2), then most recently updated
```

`sort.SliceStable`, so GitHub's own `updated-desc` order survives as the
tiebreak.

`Section.Total`, `.Active` and `.Hot` are counted **before** the `onlyActive`
filter drops anything, so a filtered tab can still tell you how many pull
requests it is not showing.

## Partial failure

`BatchSearch` returns whatever aliases resolved alongside its error. If some
results came back, the board is built from them and the error is attached as
`Board.Warning`; only a total failure returns an error to the caller. One
unreadable organization must not blank the other nine tabs.
