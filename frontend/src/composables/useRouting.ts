/**
 * Which section the page is showing, kept in the address bar.
 *
 * The app is still one page with no router library: there is one piece of
 * state, the id of the visible section, and the path is how it is written down.
 * That is enough to give the dashboard and each queue a link you can bookmark,
 * reload, or send to somebody, and to make the back button do what it looks
 * like it should.
 *
 * The mapping itself is in lib/routes.ts. This half owns the ref and the
 * history API, and it is a module-level ref like the rest of the app's shared
 * state -- see composables/useTheme.ts.
 */

import { ref } from 'vue'

import { pathForSection, sectionForPath } from '../lib/routes'

function readPath(): string {
  if (typeof location === 'undefined') return 'dashboard'
  return sectionForPath(location.pathname) ?? 'dashboard'
}

const active = ref(readPath())

export function useRouting() {
  return {
    /** The visible section's id. Read it; write it through `go`. */
    activeSection: active,

    /**
     * Show a section and record it.
     *
     * A click is a new history entry, so back returns you to where you were. A
     * correction is not: `replace` is for the cases where the app decides the
     * path is wrong -- an unknown path on arrival, or a followed team that has
     * since been dropped -- and a back button that walked into those again
     * would be a trap.
     */
    go(id: string, replace = false) {
      const path = pathForSection(id)
      if (typeof history !== 'undefined' && path !== location.pathname) {
        if (replace) history.replaceState(null, '', path)
        else history.pushState(null, '', path)
      }
      active.value = id
    },

    /** Follows the back and forward buttons. Returns its own teardown. */
    startRouting() {
      if (typeof window === 'undefined') return () => {}
      const onPop = () => (active.value = readPath())
      window.addEventListener('popstate', onPop)
      return () => window.removeEventListener('popstate', onPop)
    },
  }
}
