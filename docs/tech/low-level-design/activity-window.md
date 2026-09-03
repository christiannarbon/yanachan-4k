# Activity window

`backend/internal/board/window.go`

The window decides what counts as recent. Everything newer than `Cutoff` is new
activity; everything older is history.

## The business-day rule

```go
ResolveWindow(now time.Time, hoursOverride int) Window
```

With `hoursOverride <= 0`:

| `now.Weekday()` | Days back | Kind | Label |
| --- | --- | --- | --- |
| Tue–Fri | 1 | `fixed` | `last 24h` |
| Sat | 1 | `business-day` | `since last business day (Fri)` |
| Sun | 2 | `business-day` | `since last business day (Fri)` |
| Mon | 3 | `business-day` | `since last business day (Fri)` |

The cutoff is `now.AddDate(0, 0, -days)` — calendar days, not
`24 * days` hours, so a DST transition inside the window does not shift it by an
hour.

Tue–Fri report `Kind: fixed` even though no override was given. That is
deliberate: the *kind* describes what the sentence should say, and on a Tuesday
the honest sentence is "last 24h", not "since last business day".

## The override

A positive `hoursOverride` replaces the rule entirely: `Kind: fixed`, `Hours`
as given, cutoff at `now.Add(-hours)`.

It arrives from two places, in this order of precedence:

1. `GET /api/board?hours=N` — a query parameter, accepted for `0 < N <= 720`.
2. `settings.WindowHours` — the settings dialog, same bound, `0` meaning the rule.

720 is `state.MaxWindowHours`, 30 days. The bound exists because
hours-to-`Duration` multiplication overflows outright somewhere past 292 years,
and nothing about this dashboard is useful at that range anyway.

## The payload

```go
type Window struct {
    Kind   string    // "fixed" | "business-day"
    Label  string    // the same sentence, in English
    Hours  int
    Cutoff time.Time
    Now    time.Time
}
```

`Kind` and `Hours` are the machine-readable pair; the frontend builds its own
sentence from them so the window reads correctly in Japanese. `Label` is kept so
a client that does not translate still has something to print. If you add a
window kind, add it to the frontend's `windowLabel` too — the type system will
not catch that one, because `Kind` is a string.

## Tests

`TestResolveWindow`, in `board/build_test.go`, walks every weekday and both
branches. It is the
cheapest test in the repository and the one most likely to catch a refactor that
looked harmless.
