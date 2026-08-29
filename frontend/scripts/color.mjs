/** Colour maths shared by the theme generator: sRGB <-> HSL, WCAG contrast. */

export function hexToRgb(hex) {
  const h = hex.replace('#', '')
  return [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16))
}

export function rgbToHex([r, g, b]) {
  const c = (n) => Math.max(0, Math.min(255, Math.round(n))).toString(16).padStart(2, '0')
  return `#${c(r)}${c(g)}${c(b)}`
}

export function rgbToHsl([r, g, b]) {
  r /= 255; g /= 255; b /= 255
  const max = Math.max(r, g, b), min = Math.min(r, g, b)
  const l = (max + min) / 2
  if (max === min) return [0, 0, l]
  const d = max - min
  const s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
  let h
  if (max === r) h = ((g - b) / d + (g < b ? 6 : 0))
  else if (max === g) h = (b - r) / d + 2
  else h = (r - g) / d + 4
  return [h * 60, s, l]
}

export function hslToRgb([h, s, l]) {
  h = ((h % 360) + 360) % 360
  if (s === 0) return [l * 255, l * 255, l * 255]
  const q = l < 0.5 ? l * (1 + s) : l + s - l * s
  const p = 2 * l - q
  const f = (t) => {
    t = ((t % 1) + 1) % 1
    if (t < 1 / 6) return p + (q - p) * 6 * t
    if (t < 1 / 2) return q
    if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6
    return p
  }
  return [f(h / 360 + 1 / 3), f(h / 360), f(h / 360 - 1 / 3)].map((v) => v * 255)
}

export const hexToHsl = (hex) => rgbToHsl(hexToRgb(hex))
export const hslToHex = (hsl) => rgbToHex(hslToRgb(hsl))

/** WCAG relative luminance. */
export function luminance(hex) {
  const [r, g, b] = hexToRgb(hex).map((v) => v / 255)
  const f = (c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4)
  return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b)
}

/** WCAG contrast ratio, 1..21. */
export function contrast(a, b) {
  const la = luminance(a), lb = luminance(b)
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05)
}

/** Worst contrast of `fg` against every background in `bgs`. */
export const minContrast = (fg, bgs) => Math.min(...bgs.map((b) => contrast(fg, b)))

/**
 * Walks a colour's lightness until it clears `target` against every background,
 * keeping hue and saturation so the painting's character survives the fix.
 * Tries both directions and returns whichever gets closest.
 */
export function fitContrast(hex, bgs, target) {
  if (minContrast(hex, bgs) >= target) return hex
  const [h, s] = hexToHsl(hex)
  let best = hex, bestScore = minContrast(hex, bgs)
  for (const dir of [-1, 1]) {
    const [, , l0] = hexToHsl(hex)
    for (let step = 1; step <= 100; step++) {
      const l = l0 + dir * step * 0.01
      if (l < 0 || l > 1) break
      const cand = hslToHex([h, s, l])
      const score = minContrast(cand, bgs)
      if (score > bestScore) { best = cand; bestScore = score }
      if (score >= target) return cand
    }
  }
  return best
}

/**
 * Blends `a` toward `b`, then pulls back to `target` contrast. Used for muted
 * and faint text, which should sit between the body text and its surface.
 */
export function mix(a, b, t) {
  const ra = hexToRgb(a), rb = hexToRgb(b)
  return rgbToHex(ra.map((v, i) => v + (rb[i] - v) * t))
}

/** The text colour in `candidates` with the best worst-case contrast. */
export function pickText(candidates, bgs, target) {
  let best = null, bestScore = -1
  for (const c of candidates) {
    if (!c) continue
    const score = minContrast(c, bgs)
    if (score >= target) return { hex: c, score, repaired: false }
    if (score > bestScore) { best = c; bestScore = score }
  }
  const fixed = fitContrast(best, bgs, target)
  return { hex: fixed, score: minContrast(fixed, bgs), repaired: true }
}

/**
 * Rough perceptual distance between two colours, as a plain Euclidean
 * distance in sRGB weighted toward green. Good enough to answer the only
 * question asked of it: would a reader take these two dots for the same
 * colour? Below roughly 60 the answer is yes.
 */
export function separation(a, b) {
  const [r1, g1, b1] = hexToRgb(a)
  const [r2, g2, b2] = hexToRgb(b)
  return Math.sqrt(2 * (r1 - r2) ** 2 + 4 * (g1 - g2) ** 2 + 3 * (b1 - b2) ** 2) / Math.sqrt(9)
}
