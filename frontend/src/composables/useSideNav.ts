/**
 * Side navigation state: which headings are folded, and whether the drawer is
 * showing on a narrow window.
 *
 * The navigation lists every followed team and organization, so on a busy
 * account it is the longest thing on the page. Folding a heading away is how
 * that list is made short again, and the choice is worth keeping -- fold the
 * organizations once and they stay folded on the next visit.
 *
 * The folded set is a module-level ref with a `localStorage` mirror, the same
 * shape the theme, the locale and the repository folds use; see
 * composables/useRepoGroups.ts. The drawer is not persisted: it belongs to the
 * window it was opened in.
 */

import { ref, watch } from 'vue'

/** The headings the navigation can fold. Fixed, so the key stays readable. */
export type NavGroupId = 'queues' | 'teams' | 'orgs'

const FOLDED_KEY = 'yana.navGroups'

function readStored(): NavGroupId[] {
  try {
    const raw = localStorage.getItem(FOLDED_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      return parsed.filter((v): v is NavGroupId => v === 'queues' || v === 'teams' || v === 'orgs')
    }
  } catch {
    // Private windows and blocked site data throw; so does a value written by
    // some other version of this key. Starting fully expanded is the safe read.
  }
  return []
}

const folded = ref(new Set<NavGroupId>(readStored()))

watch(
  folded,
  (set) => {
    try {
      localStorage.setItem(FOLDED_KEY, JSON.stringify([...set]))
    } catch {
      // Persisting is a convenience; the folds still hold for this session.
    }
  },
  { deep: true },
)

/** Whether the drawer is showing. Only a narrow window ever hides the rail. */
const drawerOpen = ref(false)

export function useSideNav() {
  return {
    drawerOpen,

    isFolded: (group: NavGroupId) => folded.value.has(group),

    toggleGroup(group: NavGroupId) {
      if (folded.value.has(group)) folded.value.delete(group)
      else folded.value.add(group)
    },
  }
}
