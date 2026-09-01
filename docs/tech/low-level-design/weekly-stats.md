# Weekly stats

`backend/internal/stats/build.go`, `types.go`

Where `board` answers "what needs me right now", this answers "what did I get
done". The two share a client and a login and nothing else.

## The window: whole local days

```go
ResolveWeek(now time.Time, days int) Week
```

`Since` is midnight local time, `days - 1` days before today; `Until` is `now`.
Whole days, not a rolling 168 hours — the chart draws a column per day and
labels it with a weekday, and a window starting at 14:07 last Sunday would put
two half-Sundays at the ends and make the first and last columns lie about
themselves.

`days` is clamped: below 1 becomes `DefaultDays` (7), above `MaxDays` (90)
becomes 90. The API accepts `?days=N` in `1..90`; the tab asks for 7.

`Since` is formatted `RFC3339` into the search qualifiers, so the day boundaries
GitHub filters on are the ones the chart draws.

The window is the only bound on the request. `Request` carries no page size,
because a total has to cover all of its window — see
[the week pages, the board does not](graphql-batching.md#the-week-pages-the-board-does-not).

## The four searches

One batched round trip over the lighter `PRStat` fragment, paged to the end of
the window:

| Alias | Query |
| --- | --- |
| `opened` | `is:pr author:me created:>=SINCE` |
| `merged` | `is:pr author:me merged:>=SINCE` |
| `closed` | `is:pr author:me is:unmerged closed:>=SINCE` |
| `reviewed` | `is:pr reviewed-by:me -author:me updated:>=SINCE` |

Each alias is paged until GitHub runs out of results for it, so a week busier
than one page is counted rather than sampled. Every result is then re-checked
against the window in Go with `inWeek`, which excludes both edges. GitHub's
`>=SINCE` is a coarse filter; the exact bucket is decided here, in the server's
zone.

`reviewed` searches on `updated:` rather than a review timestamp, because there
is no search qualifier for "reviewed on". A pull request you reviewed three
weeks ago that was updated yesterday comes back and is then dropped by
`countReviews` returning no in-week days.

## Counting reviews

```go
countReviews(pr, me, week) (written, approvals int, days []string)
```

Walks the PR's reviews and keeps only yours, only inside the window, and only
those that are **not** `PENDING` — a review still open in somebody's browser has
not been submitted and is not work delivered.

`written` counts submitted reviews; `approvals` counts the `APPROVED` subset;
`days` is the **distinct** day keys, which is why the tile and the chart can
disagree:

- `Stats.Reviewed` is incremented once per pull request that had any in-week
  review — **distinct branches**.
- `Day.Reviewed` is incremented once per distinct day that branch was reviewed
  on — so a branch you went back and forth on across two days contributes 1 to
  the tile and 2 across the strip.

Both are true. They answer different questions, and this is the single most
common "the numbers don't add up" report.

## Day buckets

`dayKey` formats the timestamp as `YYYY-MM-DD` **in the window's zone**, and
`emptyDays` pre-seeds one `Day` per column so a day with no activity is a zero
column rather than a missing one. `TestDayKeyBucketsInTheWindowsZone` covers the
non-UTC case, which is the one that breaks if somebody reaches for `.UTC()`.

## Rhythm

```go
rhythm(days) (active, streak int)
```

`active` is how many columns are non-zero. `streak` is the run of non-zero days
ending at the last column — **allowing today to be empty**, because the
dashboard is often open at 09:00 on a day that has not happened yet, and
breaking the streak for that would be a lie about yesterday.

## Superlatives

Taken over merged pull requests only: an open branch has no time-to-merge and
its diff is still moving.

| Highlight | Rule |
| --- | --- |
| Fastest merge | smallest `MergedAt - CreatedAt` |
| Biggest merge | largest `Additions + Deletions` |
| Busiest repository | most pull requests touched; ties break by name |

Every highlight is a pointer and may be nil. A week with nothing merged has no
fastest merge, and the frontend leaves the row out rather than printing a zero.

## The calorie total

```go
Kcal = Opened*200 + Merged*400 + Closed*100 + ReviewsWritten*150 + Approvals*50
```

The weights are `const` in `types.go` and are the only place the number is
decided. They are arbitrary by definition, so they are at least arbitrary in one
place, and ordered the way the work is: shipping beats opening, reviewing
somebody else's branch counts, and a pull request you closed unmerged still cost
you the afternoon.

Note that `ReviewsWritten` is used, not `Reviewed` — the calorie figure rewards
each submitted review, while the tile counts branches.

## Partial failure

Same contract as the board: if some aliases resolved, the payload is built from
them and the error becomes `Stats.Warning`. Only a total failure returns an
error.
