# Theming system

Ten palettes, one semantic vocabulary, and a generator that will not let a theme
ship an unreadable pair of colours.

## The files

```
frontend/src/styles/theme.css                 what a theme does not change: spacing, motion, fallbacks
frontend/src/styles/theme-transition.css      the sweep from one palette to the next
frontend/src/styles/art-themes.css            GENERATED — one palette per theme, keyed by [data-art]
frontend/src/styles/art-themes.meta.json      names, subtitles, fonts and swatches for the picker
frontend/src/styles/art-themes.audit.md       GENERATED — the contrast audit for every theme
frontend/src/styles/calorie-meter.css         the Yanami theme's own drawing (see below)
frontend/scripts/gen-art-themes.mjs           regenerates art-themes.css, the meta and the audit
frontend/scripts/color.mjs                    the colour maths the generator uses
```

Three of those are generated. Editing `art-themes.css` by hand works until the
next regeneration silently reverts it.

## The vocabulary

Components use semantic custom properties, never a colour literal and never a
theme name:

| Token family | Role |
| --- | --- |
| `--bg`, `--bg-sunken`, `--panel`, `--panel-raised` | surfaces |
| `--text`, `--text-muted`, `--text-faint` | type, in descending contrast |
| `--fact`, `--fact-soft`, `--on-fact`, `--on-fact-soft` | the primary; the "active" border |
| `--dim`, `--dim-soft`, `--on-dim`, `--on-dim-soft` | attention; the "needs you" border |
| `--danger`, `--warning`, `--info`, `--ok` and their soft/on pairs | status |
| `--border`, `--border-strong`, `--edge`, `--edge-strong` | rules and outlines |
| `--accent`, `--accent-hover`, `--accent-contrast`, `--focus-ring` | interactive |
| `--radius*`, `--font-display`, `--font-sans` | shape and type |

`theme.css` holds the parts a theme does not touch — the spacing scale, the
motion tokens (`--dur`, `--ease`), the font fallback stacks — plus a complete
default palette, so a missing `[data-art]` still renders a coherent page.

`art-themes.css` redefines only the palette, under `[data-art="<id>"]`.

## The generator

```sh
cd frontend && node scripts/gen-art-themes.mjs <upstream-checkout>/themes
```

Eight painting palettes come from the M3 token sets in
[art_inspired_design_system_for_AI](https://github.com/peiqingzhang/art_inspired_design_system_for_AI).
Studio Paper and Yanami are defined in the generator itself and go through the
**identical** pipeline, so every theme carries the same guarantees.

The mapping is not a blind copy. Some upstream palettes span a wider lightness
range than any single text colour can serve — Matisse's ramp runs from a deep
red surface to a pale salmon container, where black reaches only 3.8:1 at one
end and white 2.2:1 at the other. So every derived colour is checked against the
surfaces it will **actually** sit on and nudged in lightness, holding hue and
saturation, until it clears its target. The painting still sets the character;
the generator only guarantees it stays readable.

Two checks beyond contrast:

- **Separation** — a perceptual distance between colours that must stay
  distinguishable, notably `--fact` against `--dim` (the two card borders) and
  `--danger` against `--warning`. Below roughly 45 two colours read as one.
- **Fallback species** — the generator knows which display faces are serifs, so
  a theme whose webfont never lands falls back to something of the same species.
  Space Grotesk falling back to Georgia changes the whole page's voice.

Every worst-case figure is written to `art-themes.audit.md`, committed. A
theme's accessibility is a file you can read rather than a claim in a README.

## The picker

`art-themes.meta.json` drives it: `id` (matching `[data-art]`), `name`,
`subtitle`, `kind` (`ode` | `house` | `painting`), the `fonts` it asks for, and
a `swatch` — the theme's fact colour.

`composables/useTheme.ts` holds the active id, writes it to `[data-art]` on the
document root, mirrors it to `localStorage` under `yana.art`, and loads the
theme's webfonts **on first selection**. Nine of the ten cost nothing until
somebody tries them.

There is one axis. The app is light-only by design, so a theme is a single
palette rather than a light/dark pair, and there is no `prefers-color-scheme`
block anywhere.

## The sweep

A palette change is a whole-page repaint, which is instant and reads as a
glitch. `composables/useThemeTransition.ts` runs it inside a view transition
instead: the browser snapshots the old page, the attribute changes, and the new
page wipes in from the left behind a gradient mask three viewports wide, slid
from one end to the other. The animation is in `theme-transition.css`; the
duration and easing are `--dur-sweep` and `--ease-sweep` in `theme.css`.

Every rule is behind `[data-theme-sweep]`, set on the document root only for the
length of a theme change, so a view transition added anywhere else later gets
the browser's own cross-fade rather than this one.

Two ways out, both silent: a browser without `startViewTransition`, and a
`prefers-reduced-motion: reduce` that is checked at the moment of the change.
Either one swaps the palette outright, which is what the app did before any of
this existed. The picker also closes itself a tick before it sets the theme, so
an open menu is in neither snapshot.

## Adding a theme

1. Add its palette to the generator — either as an upstream directory or, for a
   hand-made palette, in the same shape as Studio Paper.
2. Add a title/subtitle entry if the slug is not presentable.
3. Run the generator.
4. Read the new row in `art-themes.audit.md`. If a figure is below target the
   generator will have nudged the colour; if the *separation* columns are low,
   the palette genuinely does not have two distinguishable accents and needs a
   different pair chosen.
5. Commit all three generated files together.

No component changes. If a theme needs one, something has escaped the vocabulary.

## The one exception

`calorie-meter.css` is the only file in the project where a component-level rule
knows a theme's name, and it is deliberate.

What makes the calorie meter page recognisable is not its palette but its
*drawing*: totals set as Quantico numerals in yellow tiles, white cards ringed in
a navy outline inside the corner radius, a vermillion cap on the figure that
matters, a flat yellow marker behind the heading. Tokens cannot express any of
that — there is no `--tile-that-is-yellow-because-of-a-website`.

So those four shapes live in one file, every rule behind
`[data-art="yanami-calorie-meter"]`, inert under every other theme.

Nothing in it is load-bearing: delete the file and the theme is still a
complete, legible palette — it just stops being an ode. That is the test for
whether a future exception is allowed to exist.
