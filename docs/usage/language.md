# Language

The whole interface is available in English and Japanese. The switch sits in the
top corner beside the [theme picker](themes.md).

The first visit follows the browser's language preferences. The choice is then
remembered per browser under `yana.locale`.

## What is not translated

Errors relayed from the backend or from GitHub arrive in English in both
languages. A GraphQL error message is GitHub's text, and translating it would
mean paraphrasing an error — worse than reading it as sent.

## Adding or changing copy

Copy lives in two catalogs, `frontend/src/i18n/en.ts` and `ja.ts`. English is the
source of truth: `ja.ts` is typed as its shape, so `make test` fails if a key is
added to one and not the other. Strings are read as property access
(`t.board.refresh`), not by key string, so a typo is a type error too.

The mechanics, including how a sentence with inline markup keeps its grammar in
both languages, are in [internationalization](../tech/low-level-design/i18n.md).
