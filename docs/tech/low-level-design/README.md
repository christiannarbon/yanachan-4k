# Low level design

One document per rule set, each one pointing at the file that implements it.

## Contracts

- [API reference](api-reference.md) — the HTTP surface between the two halves.
- [Configuration reference](configuration-reference.md) — every environment
  variable, its default, and what reads it.

## The logic ported from the shell script

- [Activity window](activity-window.md) — what "recent" means, and why it
  changes on a Monday.
- [Board classification](board-classification.md) — the event timeline, the
  `Reply` / `New` branches, bot detection, and one inherited quirk that is
  reproduced on purpose.

## The logic that is new

- [Weekly stats](weekly-stats.md) — whole-day buckets, distinct-PR counting,
  the superlatives, and the calorie weights.
- [GraphQL batching](graphql-batching.md) — how N tabs stay one round trip.

## Everything else

- [State persistence](state-persistence.md) — two JSON files, `0600`, written
  atomically, tolerant of corruption.
- [Theming system](theming-system.md) — the token vocabulary, the generator
  that produces the palettes and their contrast audit.
- [Internationalization](i18n.md) — English as the source of truth, Japanese
  type-checked against it.
