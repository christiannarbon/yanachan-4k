/**
 * The map between a section and the path that shows it.
 *
 * Pure string work, kept apart from the history API so the mapping can be read
 * in one place: composables/useRouting.ts is the half that talks to the browser.
 *
 *   /dashboard                          your week
 *   /prs/mine                           your open pull requests
 *   /prs/review                         reviews requested from you
 *   /prs/team/acme/platform             a followed team
 *   /prs/org/acme                       a followed organization
 *   /settings                           settings
 *
 * A section id arrives from the backend as `mine`, `review`, `team:org/slug` or
 * `org:login`, and the colon is the only part that needs translating: a team's
 * ref already carries the slash a path wants.
 */

/** Where the app goes when the path does not name anything it knows. */
export const DEFAULT_PATH = '/dashboard'

/** The path that shows a section, ready for history.pushState. */
export function pathForSection(id: string): string {
  if (id === 'dashboard') return '/dashboard'
  if (id === 'settings') return '/settings'
  if (id.startsWith('team:')) return `/prs/team/${id.slice('team:'.length)}`
  if (id.startsWith('org:')) return `/prs/org/${id.slice('org:'.length)}`
  // mine, review, and any kind added later.
  return `/prs/${encodeURIComponent(id)}`
}

/**
 * The section a path names, or null if it names nothing. Whether that section
 * exists on the board is the caller's business; this only reads the path.
 */
export function sectionForPath(pathname: string): string | null {
  const path = pathname.replace(/\/+$/, '')
  if (path === '' || path === '/dashboard') return 'dashboard'
  if (path === '/settings') return 'settings'
  if (!path.startsWith('/prs')) return null

  const rest = path.slice('/prs'.length).replace(/^\//, '')
  if (rest === '') return null
  if (rest.startsWith('team/')) return `team:${rest.slice('team/'.length)}`
  if (rest.startsWith('org/')) return `org:${rest.slice('org/'.length)}`
  // A bare queue. It came through encodeURIComponent, so a stray slash here is
  // not a queue id and the path is not one this app wrote.
  if (rest.includes('/')) return null
  return decodeURIComponent(rest)
}
