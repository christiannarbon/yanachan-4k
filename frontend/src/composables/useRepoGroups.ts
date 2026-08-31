/**
 * Repository grouping for the queue lists.
 *
 * A queue that spans a dozen repositories reads as one long undifferentiated
 * column, so the entries of a section are split by repository and each group
 * gets a heading that folds it away.
 *
 * Grouping rearranges the list; it never re-ranks it. The board arrives sorted
 * with the pull requests waiting on you at the top, so a group takes the
 * position of its first entry and the entries inside it keep the order they
 * came in. The repository holding the most urgent pull request therefore still
 * leads the page.
 *
 * Which groups are folded is remembered per browser, keyed by repository and
 * shared across the tabs -- fold away the noisy repository once and it stays
 * folded on every queue and across reloads. The state lives here as a
 * module-level ref with a `localStorage` mirror, the same shape the theme and
 * the locale use; see composables/useTheme.ts.
 */

import { ref, watch } from 'vue'

import type { Entry } from '../lib/types'

/** One repository's slice of a section, in the order the section gave it. */
export interface RepoGroup {
  /** `owner/name`, as the entries carry it. */
  repo: string
  entries: Entry[]
  /** How many of `entries` need attention, for the heading's own badge. */
  hot: number
}

const COLLAPSED_KEY = 'yana.collapsedRepos'

function readStored(): string[] {
  try {
    const raw = localStorage.getItem(COLLAPSED_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed.filter((v): v is string => typeof v === 'string')
  } catch {
    // Private windows and blocked site data throw; so does a value written by
    // some other version of this key. Starting fully expanded is the safe read.
  }
  return []
}

const collapsed = ref(new Set(readStored()))

watch(
  collapsed,
  (set) => {
    try {
      localStorage.setItem(COLLAPSED_KEY, JSON.stringify([...set]))
    } catch {
      // Persisting is a convenience; the folds still hold for this session.
    }
  },
  { deep: true },
)

/**
 * Splits a section's entries by repository, preserving the order of both the
 * repositories and the entries within each one.
 */
export function groupByRepo(entries: Entry[]): RepoGroup[] {
  const groups: RepoGroup[] = []
  const byRepo = new Map<string, RepoGroup>()
  for (const entry of entries) {
    let group = byRepo.get(entry.repo)
    if (!group) {
      group = { repo: entry.repo, entries: [], hot: 0 }
      byRepo.set(entry.repo, group)
      groups.push(group)
    }
    group.entries.push(entry)
    if (entry.hot) group.hot += 1
  }
  return groups
}

export function useRepoGroups() {
  return {
    isCollapsed: (repo: string) => collapsed.value.has(repo),

    toggle(repo: string) {
      if (collapsed.value.has(repo)) collapsed.value.delete(repo)
      else collapsed.value.add(repo)
    },

    /** Folds or unfolds a whole list of repositories in one go. */
    setAll(repos: string[], fold: boolean) {
      for (const repo of repos) {
        if (fold) collapsed.value.add(repo)
        else collapsed.value.delete(repo)
      }
    },
  }
}
