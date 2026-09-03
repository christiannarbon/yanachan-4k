<script setup lang="ts">
/**
 * The board's navigation, down the left side of the page.
 *
 * It used to be a strip of tabs along the top. That reads well with three or
 * four sections and badly with twenty: a followed team or organization is
 * named by its ref, the refs are long, and the strip turned into a horizontal
 * scroller you had to drag through to reach the queue you wanted. Down the
 * side there is room for the whole list at once, and the names have somewhere
 * to go.
 *
 * The list is grouped -- queues, then teams, then organizations -- and each
 * heading is a disclosure button that folds its own group away, so an account
 * following a dozen orgs can put them out of sight and keep them one click
 * away. What is folded persists; see composables/useSideNav.ts.
 *
 * A folded heading holding the section you are on says so with its own marker,
 * so the navigation never looks like nothing is selected.
 */

import { computed } from 'vue'

import { useSideNav, type NavGroupId } from '../composables/useSideNav'
import { useI18n } from '../i18n'
import type { Section } from '../lib/types'

const props = defineProps<{ sections: Section[]; activeId: string }>()
const emit = defineEmits<{ (e: 'select', id: string): void }>()

const { t, sectionTitle } = useI18n()
const { isFolded, toggleGroup } = useSideNav()

interface NavGroup {
  id: NavGroupId
  title: string
  sections: Section[]
  /** The group's own total, for the heading badge while it is folded. */
  hot: number
}

/** The sections, bucketed by kind. An empty bucket is left out entirely. */
const groups = computed<NavGroup[]>(() => {
  const buckets: Record<NavGroupId, Section[]> = { queues: [], teams: [], orgs: [] }
  for (const section of props.sections) {
    // A kind added later lands with the queues rather than disappearing.
    if (section.kind === 'team') buckets.teams.push(section)
    else if (section.kind === 'org') buckets.orgs.push(section)
    else buckets.queues.push(section)
  }
  const titles: Record<NavGroupId, string> = {
    queues: t.value.nav.queues,
    teams: t.value.nav.teams,
    orgs: t.value.nav.orgs,
  }
  const order: NavGroupId[] = ['queues', 'teams', 'orgs']
  return order
    .filter((id) => buckets[id].length > 0)
    .map((id) => ({
      id,
      title: titles[id],
      sections: buckets[id],
      hot: buckets[id].reduce((sum, s) => sum + s.hot, 0),
    }))
})

/** Which heading holds the current section, so a folded one can still mark it. */
const activeGroup = computed(
  () => groups.value.find((g) => g.sections.some((s) => s.id === props.activeId))?.id ?? null,
)
</script>

<template>
  <nav class="nav" :aria-label="t.nav.label">
    <button
      class="nav-item nav-lead"
      :class="{ 'nav-current': activeId === 'dashboard' }"
      :aria-current="activeId === 'dashboard' ? 'page' : undefined"
      @click="emit('select', 'dashboard')"
    >
      <span class="nav-label">{{ t.sections.dashboard }}</span>
    </button>

    <section v-for="group in groups" :key="group.id" class="nav-group">
      <h2 class="nav-head">
        <button
          class="nav-toggle"
          :class="{ 'nav-toggle-current': isFolded(group.id) && activeGroup === group.id }"
          :aria-expanded="!isFolded(group.id)"
          @click="toggleGroup(group.id)"
        >
          <svg
            class="chev"
            :class="{ 'chev-open': !isFolded(group.id) }"
            viewBox="0 0 12 12"
            aria-hidden="true"
          >
            <path
              d="M4.25 2.5 L7.75 6 L4.25 9.5"
              fill="none"
              stroke="currentColor"
              stroke-width="1.6"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span class="nav-label">{{ group.title }}</span>
          <span
            v-if="isFolded(group.id) && group.hot > 0"
            class="nav-dot"
            :title="t.board.needingAttention(group.hot)"
          ></span>
        </button>
      </h2>

      <ul v-if="!isFolded(group.id)" class="nav-list">
        <li v-for="section in group.sections" :key="section.id">
          <button
            class="nav-item"
            :class="{ 'nav-current': section.id === activeId }"
            :aria-current="section.id === activeId ? 'page' : undefined"
            @click="emit('select', section.id)"
          >
            <span class="nav-label">{{ sectionTitle(section) }}</span>
            <span class="nav-count" :class="{ 'nav-count-hot': section.hot > 0 }">
              {{ section.total }}
            </span>
            <span
              v-if="section.hot > 0"
              class="nav-dot"
              :title="t.board.needingAttention(section.hot)"
            ></span>
          </button>
        </li>
      </ul>
    </section>
  </nav>
</template>

<style scoped>
.nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 14px 10px 20px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 7px;
  width: 100%;
  padding: 7px 10px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 500;
  text-align: left;
}
.nav-item:hover:not(:disabled) { background: var(--bg-sunken); border-color: transparent; }
.nav-current {
  background: var(--fact-soft);
  border-color: var(--fact-line);
  color: var(--heading);
  font-weight: 600;
}
.nav-current:hover:not(:disabled) { background: var(--fact-soft); border-color: var(--fact-line); }

/* Your week leads the rail, ahead of the groups it summarises. */
.nav-lead { margin-bottom: 6px; }

.nav-group + .nav-group { margin-top: 10px; }
.nav-head { margin: 0; }

/* A heading is a label on the list below it, not another destination, so it is
   quieter than an item and keeps only the chevron to say it can be folded. */
.nav-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 5px 8px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-faint);
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  text-align: left;
}
html[lang="ja"] .nav-toggle { letter-spacing: 0; text-transform: none; }
.nav-toggle:hover:not(:disabled) { background: var(--bg-sunken); border-color: transparent; }
/* Folded over the section you are on: the marker the hidden item would carry. */
.nav-toggle-current { color: var(--heading); }

.chev {
  flex: none;
  width: 11px;
  height: 11px;
  transition: transform var(--dur) var(--ease);
}
.chev-open { transform: rotate(90deg); }

.nav-list { margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 2px; }

/* Refs are long -- org/a-fairly-long-team-slug -- so the name takes the width
   that is left and gives up its tail rather than widening the rail. */
.nav-label { flex: 1 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.nav-count {
  flex: none;
  min-width: 20px;
  padding: 0 6px;
  border-radius: var(--radius-full);
  background: var(--bg-sunken);
  color: var(--text-faint);
  font-size: 11px;
  font-weight: 600;
  text-align: center;
}
.nav-count-hot { background: var(--dim-soft); color: var(--on-dim-soft); }
.nav-dot {
  flex: none;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--dim);
}
</style>
