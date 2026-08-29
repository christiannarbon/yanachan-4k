# Your week

The tab the app opens on. Every other tab is a queue — things waiting for you —
and this one is the opposite: a read-only look back at the last seven days.

## The four counts

| Figure | Where it comes from |
| --- | --- |
| Opened | `author:you created:>=` the start of the window |
| Merged | `author:you merged:>=` |
| Closed | `author:you is:unmerged closed:>=` |
| Reviewed | `reviewed-by:you -author:you`, then filtered to reviews you actually submitted inside the window |

Under those: reviews written, approvals given, repositories touched, and the
lines and files of everything merged. Then a strip per metric showing the seven
days on one shared scale, and the week's superlatives — fastest merge, biggest
merge, busiest repository, busiest day.

## The window is whole days

The window is whole local days ending with today, not a rolling 168 hours.

The chart draws a column per day and labels it with a weekday, so a window that
began at 14:07 last Sunday would put two half-Sundays at its ends and make the
first and last columns lie about themselves. Whole days, so a Tuesday column is
a Tuesday.

`GET /api/stats?days=N` takes anything from 1 to 90; the tab asks for 7.

## Two details before you read the numbers closely

- **Reviewed counts pull requests; the chart counts them per day.** The tile is
  distinct branches you reviewed. A strip's column is distinct branches you
  reviewed *that day*. Go back and forth on one branch across two days and the
  strip sums to one more than the tile. Both are true — they answer different
  questions.
- **Only submitted reviews count.** A review still open in your browser is
  `PENDING` to GitHub and is not work delivered, so it is skipped.

## The calorie figure

The hero number is a calorie total, because `4K` is a calorie count and that is
what this repository is named after.

| Event | kcal |
| --- | --- |
| Pull request opened | 200 |
| Pull request merged | 400 |
| Pull request closed unmerged | 100 |
| Review written | 150 |
| Approval given | 50 |

They are arbitrary, deliberately: shipping beats opening, reviewing somebody
else's branch counts for something, and a branch you closed unmerged still cost
you the afternoon. Under the Yanami theme the total is set in the calorie
meter's own yellow tile.

The weights live in one place, `backend/internal/stats/types.go` — change them
there and the whole dashboard follows.

## How the chart is drawn

Three strips, one per metric, over one shared scale, all in the theme's primary
(`--fact`).

That is a deliberate choice rather than a stacked bar. A stack would need three
hues that stay tellable apart in all ten themes, including under colour-vision
deficiency; splitting the series into a strip each removes the question. Each
strip is a single series, so identity comes from its own label rather than from
a colour, and every bar can use the one hue the contrast audit has already
cleared on that surface.

Hovering a day lights its column in all three strips and writes the figures into
the chart's heading. The same numbers are in a table behind the plot, hidden
visually and read out by assistive technology.

The reduction behind all of this is documented in
[weekly stats](../tech/low-level-design/weekly-stats.md).
