/**
 * Generates src/styles/art-themes.css.
 *
 * Three kinds of source feed this. Eight paintings come from the upstream M3
 * token sets in github.com/peiqingzhang/art_inspired_design_system_for_AI.
 * "Studio Paper" is the app's own original palette and "Yanami Anna" is the
 * one this app is named for, taken from the official 負けヒロインが多すぎる！
 * calorie meter page. Both are defined below and put through exactly the same
 * pipeline, so every theme carries the same guarantees.
 *
 * The app's components speak a small semantic vocabulary (--panel, --text,
 * --fact, --edge ...). Rather than rewrite every component in M3 role names,
 * this maps each theme's M3 tokens onto that vocabulary once and writes the
 * result as plain custom properties, selected by [data-art].
 *
 * Light only. There is no dark variant and no prefers-color-scheme block: the
 * app is light-mode by design, so a theme is one palette rather than two.
 *
 * The mapping is not a blind copy. Some upstream palettes span a wider
 * lightness range than any single text colour can serve -- Matisse's ramp runs
 * from a deep red surface to a pale salmon container, where black reaches only
 * 3.8:1 on one end and white 2.2:1 on the other. So every derived colour is
 * checked against the surfaces it will actually sit on and nudged in lightness
 * (hue and saturation held) until it clears its target. The painting still sets
 * the character; the generator only guarantees it stays readable.
 *
 * Usage:  node scripts/gen-art-themes.mjs <path-to-upstream>/themes
 */

import { readFileSync, writeFileSync, readdirSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  contrast, fitContrast, hexToHsl, hslToHex, minContrast, mix, pickText, separation,
} from './color.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const SRC = process.argv[2]
if (!SRC || !existsSync(SRC)) {
  console.error('usage: node scripts/gen-art-themes.mjs <upstream>/themes')
  process.exit(1)
}

/** Human-readable names; the directory slugs are not presentable as-is. */
const TITLES = {
  'cezanne-mont-sainte-victoire': ['Cézanne', 'Mont Sainte-Victoire'],
  'hokusai-great-wave': ['Hokusai', 'The Great Wave'],
  'hopper-nighthawks': ['Hopper', 'Nighthawks'],
  'matisse-red-studio': ['Matisse', 'The Red Studio'],
  'monet-water-lilies': ['Monet', 'Water Lilies'],
  'vangogh-green-wheat': ['Van Gogh', 'Green Wheat Fields'],
  'vangogh-irises': ['Van Gogh', 'Irises'],
  'wang-ximeng-rivers-mountains': ['Wang Ximeng', 'A Thousand Li of Rivers and Mountains'],
}

/**
 * Which display faces are serifs, so a theme falls back to something of the
 * same species if its webfont never lands. Getting this wrong is not cosmetic:
 * Space Grotesk falling back to Georgia changes the whole page's voice.
 */
const SERIF_FACES = new Set([
  'Bitter', 'Cormorant Garamond', 'Lora', 'Merriweather', 'Noto Serif Display',
  'Noto Serif JP', 'Playfair Display', 'Source Serif 4',
])

/**
 * The app's original palette, kept as a first-class theme.
 *
 * Warm paper, a teal fact colour and a burnt-orange dimension colour. It is
 * expressed here in the same M3 roles the paintings use so it goes through the
 * identical contrast and separation passes rather than being hand-written into
 * the generated stylesheet.
 */
const STUDIO_PAPER = {
  id: 'studio-paper',
  name: 'Studio Paper',
  subtitle: 'Teal & burnt orange',
  kind: 'house',
  colors: {
    primary: '#0f766e',
    'on-primary': '#ffffff',
    secondary: '#b45309',
    tertiary: '#7c2d12',
    error: '#b91c1c',

    // Warm paper. `surface` is the page and the graph canvas; the containers
    // run lighter to panels, with `highest` deliberately darker than the page
    // so a hovered row on a white panel still reads as recessed.
    surface: '#f6f4f0',
    'on-surface': '#1c1917',
    'surface-variant': '#e3dfd8',
    'on-surface-variant': '#475569',
    'surface-container-highest': '#f2f0ec',
    'surface-container-high': '#f6f4f0',
    'surface-container': '#faf9f7',
    'surface-container-low': '#ffffff',
    'surface-container-lowest': '#ffffff',

    outline: '#cec8be',
    'outline-variant': '#e3dfd8',
    'on-primary-container': '#1c1917',
    'inverse-on-surface': '#faf9f7',
    scrim: '#000000',
    shadow: '#1c1917',
  },
  corner: {
    'extra-small': '4px', small: '7px', large: '12px', 'extra-large': '16px', full: '9999px',
  },
  fonts: { brand: 'Inter', plain: 'Inter' },
}

/**
 * Yanami Anna's calorie meter, kept as a first-class theme.
 *
 * This project is an ode to her, so the palette is not invented: it is lifted
 * from the official 負けヒロインが多すぎる！ site's own custom properties, where
 * the committee's secret calorie survey lives --
 * makeine-anime.com/special/calorie_meter/. Upstream declares
 * `--color-blue: #070a7d`, `--color-orange: #ff7031`, `--color-yellow: #fff100`
 * and a pale `--color-line: #dae9f5`, and pairs "Quantico" for the big
 * kcal numerals with "Noto Sans JP" for everything else. All of that is
 * reproduced here.
 *
 * The role assignment is the interesting part. Navy is the site's structure --
 * its rules, its buttons, its 2px card outlines -- so it takes `primary`, and
 * becomes this app's --fact and --accent. Yanami's own vermillion is what the
 * page shouts with: the rotated 総摂取カロリー cap sitting on the corner of the
 * white total card. So it takes `secondary` and lands on --dim, the colour this
 * app paints on a pull request that needs you. Her colour, on the thing that is
 * asking for attention. The sky blue from the site's header gradient is left as
 * the third role.
 *
 * Like Studio Paper this is expressed in M3 roles rather than written straight
 * into the stylesheet, so it goes through the identical contrast and separation
 * passes the paintings do.
 */
const YANAMI_CALORIE_METER = {
  id: 'yanami-calorie-meter',
  name: 'Yanami Anna',
  subtitle: 'Calorie Meter',
  kind: 'ode',
  colors: {
    primary: '#070a7d',
    'on-primary': '#ffffff',
    secondary: '#ff7031',
    tertiary: '#4ac1f0',
    error: '#dd4f4e',

    // Washed sky. The special page floats white cards over a pale blue deco
    // band, so the page and the canvas take the sky and the panels stay the
    // flat white of the calorie table itself.
    surface: '#eef6fd',
    'on-surface': '#070a7d',
    'surface-variant': '#dae9f5',
    'on-surface-variant': '#98a9b3',
    'surface-container-highest': '#e2eefa',
    'surface-container-high': '#eef6fd',
    'surface-container': '#f7fbfe',
    'surface-container-low': '#ffffff',
    'surface-container-lowest': '#ffffff',

    outline: '#8fb4d8',
    'outline-variant': '#dae9f5',
    'on-primary-container': '#070a7d',
    'inverse-on-surface': '#f7fbfe',
    scrim: '#070a7d',
    shadow: '#070a7d',
  },
  // The total card, the episode tabs and the kcal digit tiles are all drawn at
  // a 10px radius upstream; the rest of the scale is built around it.
  corner: {
    'extra-small': '5px', small: '10px', large: '14px', 'extra-large': '20px', full: '9999px',
  },
  fonts: { brand: 'Quantico', plain: 'Noto Sans JP' },
}

/**
 * Contrast floors. Body text targets AA; fills and rules only need to be seen.
 * Entity roles sit above the 3:1 non-text minimum: a fact node is a ten-pixel
 * square on a busy canvas, and at the bare minimum it merely fails to vanish
 * rather than actually reading as a fact.
 */
const TARGET = {
  text: 4.5,
  muted: 4.5,
  faint: 3.0,
  role: 4.0,
  accent: 4.5,
  rule: 1.6,
  ruleStrong: 2.6,
  /** Danger is printed as text on a panel, so it owes AA. */
  statusText: 4.5,
  /**
   * Warning, info and ok are only ever chip and dot backgrounds carrying their
   * own --on- ink, so 3:1 is the honest bar. Holding them to AA forced them
   * almost black on a tinted panel, at which point an "amber" warning is no
   * longer amber -- a worse outcome than the rule was protecting against.
   */
  statusFill: 3.0,
  /** A role used as text on its own soft wash. */
  onSoft: 4.5,
}

/** Floor for entity-role chroma, so a role never collapses into grey. */
const MIN_ROLE_SAT = 0.45
/** Below roughly this, two role dots read as the same colour. */
const MIN_ROLE_SEPARATION = 45
/** Error and warning are read next to each other and must never blur. */
const MIN_STATUS_SEPARATION = 45

// --- reading upstream tokens ----------------------------------------------

/** Reads the light custom-property map from an upstream theme.css. */
function readColors(file) {
  const css = readFileSync(file, 'utf8')
  // Everything before the first @media is the light block; the dark block that
  // follows is deliberately ignored.
  const [lightBlock] = css.split('@media')
  const out = {}
  for (const m of lightBlock.matchAll(/--md-sys-color-([a-z0-9-]+):\s*(#[0-9a-fA-F]{6})/g)) {
    out[m[1]] = m[2].toLowerCase()
  }
  return out
}

const readJson = (file) => JSON.parse(readFileSync(file, 'utf8'))

/** Normalises one upstream painting directory into the shape buildTheme wants. */
function readPainting(slug) {
  const dir = join(SRC, slug, 'theme', 'tokens')
  const shape = readJson(join(dir, 'shape.tokens.json'))
  const type = readJson(join(dir, 'typography.tokens.json'))
  const corner = Object.fromEntries(
    Object.entries(shape.shape.corner)
      .filter(([k]) => !k.startsWith('$'))
      .map(([k, v]) => [k, v.$value]),
  )
  const [name, subtitle] = TITLES[slug] ?? [slug, '']
  return {
    id: slug,
    name,
    subtitle,
    kind: 'painting',
    colors: readColors(join(dir, 'theme.css')),
    corner,
    fonts: {
      brand: type.typography.ref.typeface.brand.$value,
      plain: type.typography.ref.typeface.plain.$value,
    },
  }
}

// --- helpers --------------------------------------------------------------

/**
 * Walks `hex` until it is visibly distinct from `from`, keeping hue and never
 * dropping below `target` contrast on `bgs`. Returns the best it managed if
 * the constraint cannot be met.
 */
function separate(hex, from, bgs, target, want) {
  if (separation(hex, from) >= want) return hex
  const [h, s0, l0] = hexToHsl(hex)
  let best = hex, bestScore = separation(hex, from)

  // Lightness alone runs out of room on a strongly tinted canvas: Matisse pins
  // both roles into a narrow band between the contrast floor and the salmon
  // behind them. Chroma is the other axis available, and unlike hue it costs
  // nothing in fidelity to the painting.
  for (let sat = s0; sat <= 1.0001; sat += 0.05) {
    for (const dir of [-1, 1]) {
      for (let step = 0; step <= 60; step++) {
        const l = l0 + dir * step * 0.01
        if (l < 0.05 || l > 0.95) break
        const cand = hslToHex([h, Math.min(sat, 1), l])
        if (minContrast(cand, bgs) < target) continue
        const score = separation(cand, from)
        if (score > bestScore) { best = cand; bestScore = score }
        if (score >= want) return cand
      }
    }
  }
  return best
}

/**
 * Pulls a hue inside `target ± tolerance` degrees, by the shorter way round.
 * Saturation and lightness are untouched, so the colour keeps its weight.
 */
function anchorHue(hex, target, tolerance) {
  const [h, s, l] = hexToHsl(hex)
  const delta = ((h - target + 540) % 360) - 180
  if (Math.abs(delta) <= tolerance) return hex
  return hslToHex([target + Math.sign(delta) * tolerance, s, l])
}

/** Raises a colour's saturation to at least `minS`, leaving hue and lightness. */
function vivify(hex, minS) {
  const [h, s, l] = hexToHsl(hex)
  return hslToHex([h, Math.max(s, minS), l])
}

/** A near-black carrying the source colour's hue, faintly. */
function tint(hex, lightness) {
  const [h, s] = hexToHsl(hex)
  return hslToHex([h, Math.min(s, 0.18), lightness])
}

/** Nudges a colour's lightness by `d`, used for hover states. */
function shift(hex, d) {
  const [h, s, l] = hexToHsl(hex)
  return hslToHex([h, s, Math.max(0, Math.min(1, l + d))])
}

const hexToRgb3 = (hex) =>
  [0, 2, 4].map((i) => parseInt(hex.replace('#', '').slice(i, i + 2), 16)).join(', ')

// --- the mapping ----------------------------------------------------------

/**
 * Can this ramp carry a red and an amber that both clear their bars and still
 * look like different colours? If not, no choice of status hue will save it.
 */
function statusFits(s) {
  // The same background set buildTheme fits the status colours against: they
  // are painted on panels, on the page and on recessed rows alike.
  const bgs = [s.panel, s.bg, s.sunken]
  const red = fitContrast(hslToHex([358, 0.6, 0.38]), bgs, TARGET.statusText)
  if (minContrast(red, bgs) < TARGET.statusText) return false
  const amber = fitContrast(hslToHex([45, 0.95, 0.38]), bgs, TARGET.statusFill)
  const parted = separate(amber, red, bgs, TARGET.statusFill, MIN_STATUS_SEPARATION)
  return separation(parted, red) >= MIN_STATUS_SEPARATION
}

/**
 * Chooses the surface ramp the app will paint on.
 *
 * Preferred is the full ramp, whose darkest step is the theme's own `surface`
 * -- that colour is its signature and covers the graph canvas, the largest
 * area on screen. Where including it leaves no readable text colour, the ramp
 * falls back to the container family, and then to a washed container family.
 */
function chooseSurfaces(c) {
  const full = {
    graph: c['surface'],
    bg: c['surface'],
    sunken: c['surface-container-highest'],
    panel: c['surface-container-low'],
    raised: c['surface-container-lowest'],
  }
  const containersOnly = {
    ...full,
    graph: c['surface-container-high'],
    bg: c['surface-container-high'],
  }
  const candidates = [
    c['on-surface'],
    c['on-primary-container'],
    c['inverse-on-surface'],
    tint(c['surface'], 0.08),
  ]
  const bgsOf = (s) => [s.graph, s.bg, s.sunken, s.panel, s.raised]

  const first = pickText(candidates, bgsOf(full), TARGET.text)
  if (first.score >= TARGET.text && statusFits(full)) {
    return { surfaces: full, text: first, ramp: 'full' }
  }

  const second = pickText(candidates, bgsOf(containersOnly), TARGET.text)
  if (second.score >= TARGET.text && statusFits(containersOnly)) {
    return { surfaces: containersOnly, text: second, ramp: 'containers' }
  }

  // Last resort: wash the ramp toward white.
  //
  // Matisse is the case this exists for. Its panels are a mid salmon, and
  // AA-contrast coloured text on them has to be nearly black -- which leaves
  // red and amber indistinguishable, in an app that prints "1 error, 20
  // warnings" beside each other. Reducing the contrast bar would just move the
  // failure; the ramp is what has to give. Washing it keeps every hue
  // relationship the painting set and only opens the headroom the palette was
  // short of, so the theme still reads as Matisse.
  for (let wash = 0.1; wash <= 0.65; wash += 0.05) {
    const washed = Object.fromEntries(
      Object.entries(containersOnly).map(([k, v]) => [k, mix(v, '#ffffff', wash)]),
    )
    const pick = pickText(candidates, bgsOf(washed), TARGET.text)
    if (pick.score >= TARGET.text && statusFits(washed)) {
      return { surfaces: washed, text: pick, ramp: `washed ${Math.round(wash * 100)}%` }
    }
  }
  return { surfaces: containersOnly, text: second, ramp: 'containers' }
}

function buildTheme(source) {
  const c = source.colors
  const { surfaces: sf, text: textPick, ramp } = chooseSurfaces(c)
  const textBgs = [sf.graph, sf.bg, sf.sunken, sf.panel, sf.raised]
  const chromeBgs = [sf.panel, sf.bg, sf.sunken]
  const text = textPick.hex

  // Secondary text is the body colour walked toward its surface, then pulled
  // back to AA. Deriving it this way keeps it in the theme's own family
  // instead of importing a grey that belongs to no palette.
  const muted = fitContrast(mix(text, sf.panel, 0.4), chromeBgs, TARGET.muted)
  const faint = fitContrast(mix(text, sf.panel, 0.62), chromeBgs, TARGET.faint)

  const accent = fitContrast(c['primary'], [sf.panel, sf.bg, sf.graph], TARGET.accent)
  const accentContrast = pickText(['#ffffff', '#0b0b0c', c['on-primary']], [accent], TARGET.text).hex

  // Roles and status colours are backgrounds as well as fills -- a count
  // badge, a filter chip -- and which way their own label has to go depends on
  // how light the role landed. So each ships the ink that reads on it.
  const on = (bg) => pickText(['#ffffff', '#0b0b0c'], [bg], TARGET.text).hex

  // Entity roles. These are fills and swatches beside a label, not text.
  //
  // Contrast alone is not enough. Some palettes offer a primary and secondary
  // that are barely saturated, and at ten pixels a fact node and a dimension
  // node then read as the same dot. The hues are far apart; only the chroma is
  // missing. So each role keeps its hue and is given enough saturation to
  // actually show it before contrast is fitted.
  //
  // Which M3 role plays the dimension is not fixed either. Several paintings
  // yield a primary and secondary that are all but the same colour -- Hokusai's
  // are two blues, Van Gogh's wheat fields a green and a green-teal -- while
  // the tertiary sits well clear of both. Facts keep the primary, since it is
  // also the accent and sets the theme's identity; the dimension takes
  // whichever of the remaining two is furthest from it, and the leftover
  // becomes the conformed marker, which never has to be told apart at a
  // glance. Shape still carries the distinction on the canvas -- facts are
  // square, dimensions round -- but colour should reinforce it, not fight it.
  const roleBgs = [sf.graph, sf.panel]
  const fact = fitContrast(vivify(c['primary'], MIN_ROLE_SAT), roleBgs, TARGET.role)

  const rest = ['secondary', 'tertiary']
    .map((k) => fitContrast(vivify(c[k], MIN_ROLE_SAT), roleBgs, TARGET.role))
    .sort((a, b) => separation(fact, b) - separation(fact, a))
  const [dimBase, conformed] = rest
  const dim = separate(dimBase, fact, roleBgs, TARGET.role, MIN_ROLE_SEPARATION)

  // Source models are deliberately not a domain colour; they stay neutral.
  const source_ = fitContrast(c['on-surface-variant'] ?? c['outline'], roleBgs, TARGET.role)

  // Status hues stay put across themes -- a warning that changed colour with
  // the palette would stop reading as a warning -- but their lightness is
  // fitted to whichever ramp is in play.
  const status = (h, s) => fitContrast(hslToHex([h, s, 0.38]), chromeBgs, TARGET.statusFill)
  // Warning sits further into yellow than an amber would naturally want, to
  // open the gap to danger below; at this saturation it still reads as amber.
  const warningBase = status(45, 0.95)
  const info = status(212, 0.7)
  const ok = status(145, 0.55)

  // Danger comes from M3's error role so it keeps the theme's own character,
  // but its hue is held inside a red band. Upstream harmonises error into the
  // artwork, which for Hopper and Wang Ximeng lands on an orange -- and in an
  // app whose whole job is reporting errors and warnings side by side, an
  // orange error is the same colour as an amber warning.
  const danger = fitContrast(
    vivify(anchorHue(c['error'], 358, 10), 0.6),
    chromeBgs,
    TARGET.statusText,
  )

  // Red and amber at the same lightness are nearly the same colour, and this
  // app prints "1 error, 20 warnings" side by side. Danger keeps the theme's
  // error tone; warning is the one that yields.
  const warning = separate(warningBase, danger, chromeBgs, TARGET.statusFill, MIN_STATUS_SEPARATION)

  // A "soft" is its role washed over the panel: the background of a tag or a
  // banner. The role colour itself is not always readable on it -- an amber
  // light enough to still look amber as a chip fill is too light to be text on
  // a pale amber wash. So each soft ships its own ink, the same hue walked to
  // AA, exactly as M3 pairs error-container with on-error-container.
  const soft = (hex) => mix(hex, sf.panel, 0.87)
  const inkOn = (hex, bg) => fitContrast(hex, [bg], TARGET.onSoft)

  const border = fitContrast(c['outline-variant'], [sf.panel], TARGET.rule)
  const borderStrong = fitContrast(c['outline'], [sf.panel], TARGET.ruleStrong)
  const edge = fitContrast(c['outline-variant'], [sf.graph], 1.8)
  const edgeStrong = fitContrast(c['outline'], [sf.graph], TARGET.role)
  const edgeCross = fitContrast(c['tertiary'], [sf.graph], TARGET.role)

  const shadow = hexToRgb3(c['shadow'] ?? '#000000')

  const softs = {
    fact: soft(fact), dim: soft(dim), source: soft(source_), conformed: soft(conformed),
    danger: soft(danger), warning: soft(warning), info: soft(info),
  }

  const displayFallback = SERIF_FACES.has(source.fonts.brand)
    ? 'var(--font-fallback-serif)'
    : 'var(--font-fallback-sans)'

  const vars = {
    '--radius-sm': source.corner['extra-small'],
    '--radius': source.corner['small'],
    '--radius-lg': source.corner['large'],
    '--radius-xl': source.corner['extra-large'],
    '--radius-full': source.corner['full'],
    '--font-display': `"${source.fonts.brand}", ${displayFallback}`,
    '--font-sans': `"${source.fonts.plain}", var(--font-fallback-sans)`,

    '--bg': sf.bg,
    '--bg-sunken': sf.sunken,
    '--panel': sf.panel,
    '--panel-raised': sf.raised,
    '--overlay': `rgba(${hexToRgb3(c['scrim'] ?? '#000000')}, 0.42)`,

    '--border': border,
    '--border-strong': borderStrong,

    '--text': text,
    '--text-muted': muted,
    '--text-faint': faint,

    '--fact': fact,
    '--fact-soft': softs.fact,
    '--on-fact-soft': inkOn(fact, softs.fact),
    '--on-fact': on(fact),
    '--dim': dim,
    '--dim-soft': softs.dim,
    '--on-dim-soft': inkOn(dim, softs.dim),
    '--on-dim': on(dim),
    '--source': source_,
    '--source-soft': softs.source,
    '--on-source-soft': inkOn(source_, softs.source),
    '--conformed': conformed,
    '--conformed-soft': softs.conformed,
    '--on-conformed-soft': inkOn(conformed, softs.conformed),

    '--danger': danger,
    '--danger-soft': softs.danger,
    '--on-danger-soft': inkOn(danger, softs.danger),
    '--on-danger': on(danger),
    '--warning': warning,
    '--warning-soft': softs.warning,
    '--on-warning-soft': inkOn(warning, softs.warning),
    '--on-warning': on(warning),
    '--info': info,
    '--info-soft': softs.info,
    '--on-info-soft': inkOn(info, softs.info),
    '--on-info': on(info),
    '--ok': ok,

    '--accent': accent,
    '--accent-hover': shift(accent, -0.08),
    '--accent-contrast': accentContrast,
    '--focus-ring': accent,

    '--edge': edge,
    '--edge-strong': edgeStrong,
    '--edge-cross': edgeCross,
    '--graph-bg': sf.graph,
    '--node-stroke': sf.graph,

    '--shadow-sm': `0 1px 2px rgba(${shadow}, 0.08)`,
    '--shadow': `0 2px 10px rgba(${shadow}, 0.1), 0 1px 2px rgba(${shadow}, 0.06)`,
    '--shadow-lg': `0 16px 40px rgba(${shadow}, 0.18)`,

    'color-scheme': 'light',
  }

  const audit = {
    ramp,
    text: minContrast(text, textBgs),
    muted: minContrast(muted, chromeBgs),
    faint: minContrast(faint, chromeBgs),
    accent: minContrast(accent, [sf.panel, sf.bg]),
    danger: minContrast(danger, [sf.panel, sf.bg]),
    fact: minContrast(fact, roleBgs),
    dim: minContrast(dim, roleBgs),
    factVsDim: separation(fact, dim),
    dangerVsWarning: separation(danger, warning),
    factOnSoft: contrast(inkOn(fact, softs.fact), softs.fact),
    dangerOnSoft: contrast(inkOn(danger, softs.danger), softs.danger),
    warningOnSoft: contrast(inkOn(warning, softs.warning), softs.warning),
    border: contrast(border, sf.panel),
  }

  return { vars, audit, swatch: fact }
}

// --- emit -----------------------------------------------------------------

const slugs = readdirSync(SRC)
  .filter((d) => existsSync(join(SRC, d, 'theme', 'tokens', 'theme.css')))
  .sort()

// Yanami leads -- she is what the app is named for and the default palette --
// then the house palette, then the paintings in alphabetical order.
const sources = [YANAMI_CALORIE_METER, STUDIO_PAPER, ...slugs.map(readPainting)]

const blocks = []
const meta = []
const report = [
  '# Theme contrast audit',
  '',
  'Generated by `scripts/gen-art-themes.mjs`. Every figure is the *worst* ratio',
  'across the surfaces that colour is actually painted on. Separation figures',
  'are a perceptual distance, where below roughly 45 two colours read as one.',
  '',
  '| theme | ramp | text | muted | faint | accent | danger | fact | dim | fact/soft | danger/soft | warn/soft | border | fact≠dim | danger≠warn |',
  '| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |',
]

const fmt = (o, ind) => Object.entries(o).map(([k, v]) => `${ind}${k}: ${v};`).join('\n')

for (const src of sources) {
  const { vars, audit, swatch } = buildTheme(src)

  meta.push({
    id: src.id,
    name: src.name,
    subtitle: src.subtitle,
    kind: src.kind,
    fonts: [...new Set([src.fonts.brand, src.fonts.plain])],
    swatch,
  })

  const label = src.subtitle ? `${src.name} — ${src.subtitle}` : src.name
  blocks.push(`/* ${label} */\n[data-art="${src.id}"] {\n${fmt(vars, '  ')}\n}`)

  const a = audit
  report.push(
    `| ${label} | ${a.ramp} | ${a.text.toFixed(2)} | ${a.muted.toFixed(2)} | ` +
    `${a.faint.toFixed(2)} | ${a.accent.toFixed(2)} | ${a.danger.toFixed(2)} | ` +
    `${a.fact.toFixed(2)} | ${a.dim.toFixed(2)} | ${a.factOnSoft.toFixed(2)} | ` +
    `${a.dangerOnSoft.toFixed(2)} | ${a.warningOnSoft.toFixed(2)} | ` +
    `${a.border.toFixed(2)} | ${a.factVsDim.toFixed(0)} | ${a.dangerVsWarning.toFixed(0)} |`,
  )
}

const header = `/*
 * Themes — GENERATED, do not edit by hand.
 *
 * Paintings from github.com/peiqingzhang/art_inspired_design_system_for_AI.
 * "Yanami Anna" (the default) and "Studio Paper" are defined in the generator;
 * Yanami's colours and faces come from makeine-anime.com/special/calorie_meter/.
 * Regenerate: node scripts/gen-art-themes.mjs <upstream>/themes
 *
 * One light palette per theme, selected by [data-art]. The app has no dark
 * mode, so there is no prefers-color-scheme block here by design.
 * See styles/theme.css for the base layer.
 */\n\n`

writeFileSync(join(HERE, '..', 'src', 'styles', 'art-themes.css'), header + blocks.join('\n\n') + '\n')
writeFileSync(join(HERE, '..', 'src', 'styles', 'art-themes.audit.md'), report.join('\n') + '\n')
writeFileSync(
  join(HERE, '..', 'src', 'styles', 'art-themes.meta.json'),
  JSON.stringify(meta, null, 2) + '\n',
)
console.log(`generated ${sources.length} themes`)
