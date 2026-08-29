/**
 * Checks every internal link in the Markdown documentation.
 *
 * The docs are a tree of cross-references -- docs/README.md routes into
 * usage/ and tech/, and the tech documents point at each other constantly.
 * A rename that misses one of those leaves a link that looks fine in the diff
 * and 404s for the reader, which is exactly the failure a human reviewer is
 * worst at catching.
 *
 * Two things are checked, and nothing else:
 *   1. A relative link resolves to a file that exists.
 *   2. A #fragment matches a heading in the file it points at.
 *
 * External URLs are not fetched. A link checker that hits the network is a
 * test that fails when somebody else's site is down.
 *
 * Usage:  node scripts/check-doc-links.mjs
 */

import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..')

/** Every tracked Markdown file, minus anything generated or vendored. */
const SKIP_DIRS = new Set(['node_modules', '.git', 'dist', 'bin', '.run'])

function markdownFiles(dir) {
  const out = []
  for (const name of readdirSync(dir)) {
    if (SKIP_DIRS.has(name)) continue
    const full = join(dir, name)
    if (statSync(full).isDirectory()) out.push(...markdownFiles(full))
    else if (name.endsWith('.md')) out.push(full)
  }
  return out
}

/**
 * GitHub's heading slug: lowercase, drop everything that is not a word
 * character, space or hyphen, then hyphenate the spaces. Repeats take a
 * numeric suffix, which is why this counts rather than just collecting.
 *
 * Note that each space becomes its own hyphen, so a heading with an em dash
 * in it -- which this project's headings often have -- slugs with a double
 * hyphen where the dash used to be.
 */
function headingSlugs(text) {
  const seen = new Map()
  const slugs = new Set()
  let inFence = false
  for (const line of text.split('\n')) {
    if (/^\s*```/.test(line)) {
      inFence = !inFence
      continue
    }
    if (inFence) continue
    const m = /^(#{1,6})\s+(.*)$/.exec(line)
    if (!m) continue
    const base = m[2]
      .trim()
      .toLowerCase()
      .replace(/[^\w\s-]/gu, '')
      .replace(/ /g, '-')
    const n = seen.get(base) ?? 0
    seen.set(base, n + 1)
    slugs.add(n === 0 ? base : `${base}-${n}`)
  }
  return slugs
}

const LINK = /\[[^\]]*\]\(([^)\s]+)\)/g

const problems = []
const files = markdownFiles(ROOT)

for (const file of files) {
  const text = readFileSync(file, 'utf8')
  for (const match of text.matchAll(LINK)) {
    const link = match[1]
    if (/^(https?:|mailto:|tel:)/.test(link)) continue

    const hash = link.indexOf('#')
    const path = hash === -1 ? link : link.slice(0, hash)
    const fragment = hash === -1 ? '' : link.slice(hash + 1)

    const target = path ? resolve(dirname(file), path) : file
    const where = `${relative(ROOT, file)} -> ${link}`

    if (!existsSync(target)) {
      problems.push(`${where}\n    no such file: ${relative(ROOT, target)}`)
      continue
    }
    if (!fragment) continue
    if (!target.endsWith('.md')) continue

    const slugs = headingSlugs(readFileSync(target, 'utf8'))
    if (!slugs.has(fragment)) {
      problems.push(`${where}\n    no heading in ${relative(ROOT, target)} slugs to "${fragment}"`)
    }
  }
}

if (problems.length > 0) {
  console.error(`broken documentation links (${problems.length}):\n`)
  for (const p of problems) console.error(`  ${p}\n`)
  process.exit(1)
}

console.log(`documentation links ok (${files.length} Markdown files)`)
