# GraphQL batching

`backend/internal/github/search.go`, `stats.go`, `client.go`

One refresh is two GraphQL requests, however many tabs you follow.

## The problem

The board's tabs are searches: your PRs, review-requested, reviewed-by, one per
team, one per org. Issued naively that is `3 + teams + orgs` round trips, each
with its own latency and its own rate-limit charge, and following ten teams
makes the dashboard four times slower.

## The shape of the document

`buildBatch` assembles every search as an **aliased field** on one query, all
spreading the same fragment:

```graphql
query($n:Int!, $q0:String!, $q1:String!, $q2:String!) {
  mine: search(query:$q0, type:ISSUE, first:$n) { nodes { ...PRBits } }
  req:  search(query:$q1, type:ISSUE, first:$n) { nodes { ...PRBits } }
  revd: search(query:$q2, type:ISSUE, first:$n) { nodes { ...PRBits } }
}
fragment PRBits on PullRequest { ... }
```

The searches themselves go in as **variables**, never interpolated into the
document — `$q0`, `$q1`, `$q2`, with `$n` as the shared limit. Aliases are
positional (`q0`, `q1`, …) with a caller-chosen result key (`mine`, `team0`,
`org3`).

Adding a team appends one field and one variable. The round trip count does not
move.

## Two fragments, on purpose

| Fragment | Used by | Carries |
| --- | --- | --- |
| `PRBits` | the board | comments(last:30), reviews(last:30), reviewThreads(last:20), review requests, check rollup |
| `StatBits` | the week | createdAt/closedAt/mergedAt, additions/deletions/changedFiles, reviews(last:50) |

The board asks "who said what, and when", so it drags along comments and review
threads. The week asks "what did I finish, and how big was it". Sharing one
fragment would mean every stats refresh paid for thirty comments per pull
request that nothing ever reads.

`PRBits` is the shell script's fragment, unchanged.

## Decoding

`search(type:ISSUE)` returns a union — issues come back alongside pull requests
and decode as zero values into the PR struct. Both batch functions drop any node
with `Number == 0 || URL == ""`.

Results are then sorted by `UpdatedAt` descending in Go. GitHub is already asked
for `sort:updated-desc`, but the batch functions are also used by callers that
do not set it, and the board's own ranking sorts stably on top of this.

## Partial failure is the whole point

```go
func (c *Client) Do(ctx, query, vars, out) error
```

`Do` decodes `data` **and** returns any GraphQL errors. That combination is
unusual and deliberate: GraphQL answers a partially-failing document with both,
and in this application that is the common case — one organization the token
cannot read, nine that it can.

So both batch functions return `(results, err)` with results populated, and both
callers do:

```go
results, err := cl.BatchSearch(ctx, queries, limit)
if err != nil && len(results) == 0 {
    return nil, err          // total failure
}
// ... build from what came back
if err != nil {
    b.Warning = err.Error()  // partial failure, surfaced in the UI
}
```

If you change this, the failure mode you are choosing between is "one bad org
blanks the dashboard" and "the dashboard renders with a warning". The second one
is the one people can act on.

## Client details

- One `http.Client`, 45s timeout. The server's write timeout is 90s, so a slow
  GitHub fails as a GraphQL error rather than as a truncated response.
- `Authorization: Bearer`, `User-Agent: yana-chan-4k`.
- The endpoint comes from config (`GHDASH_GRAPHQL_ENDPOINT`), which is what
  makes the client testable against a fake.

## Limits

`first: $n` is the per-search limit from settings, 1–100, default 25. It is
per tab, not in total: ten teams at 25 is 250 pull requests in one response.
`state.MaxRefs` caps teams and orgs at 200 each, because each one is another
aliased sub-query inside a single document and an unbounded list turns every
refresh into one enormous request.

There is no pagination. A tab shows the most recently updated `n`, which is what
a queue needs; going further back is what GitHub's own search page is for.
