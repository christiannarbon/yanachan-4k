<script setup lang="ts">
/**
 * One repository's pull requests under a heading that folds them away.
 *
 * The heading is a disclosure button inside its own heading element, so a
 * screen reader can jump between repositories and hears whether each one is
 * expanded. The cards are dropped from the DOM while folded rather than merely
 * hidden -- folding a repository away is how a long queue is made cheap again,
 * which only works if the rows really go -- so the button carries
 * `aria-expanded` and no `aria-controls`: the panel it would point at does not
 * exist while collapsed, and it follows the button in reading order anyway.
 */

import { useI18n } from '../i18n'
import type { Entry } from '../lib/types'
import PrCard from './PrCard.vue'

defineProps<{
  /** `owner/name`, the heading's text. */
  repo: string
  entries: Entry[]
  /** How many entries need attention, for the badge and its tooltip. */
  hot: number
  collapsed: boolean
  /** Passed through: a card words its activity line differently per section. */
  kind: string
  showUrl: boolean
  now: number
}>()

const emit = defineEmits<{ (e: 'toggle'): void }>()

const { t } = useI18n()
</script>

<template>
  <section class="group">
    <h3 class="group-head">
      <button class="group-toggle" :aria-expanded="!collapsed" @click="emit('toggle')">
        <svg class="chev" :class="{ 'chev-open': !collapsed }" viewBox="0 0 12 12" aria-hidden="true">
          <path
            d="M4.25 2.5 L7.75 6 L4.25 9.5"
            fill="none"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <span class="repo mono">{{ repo }}</span>
        <span class="count" :class="{ 'count-hot': hot > 0 }">{{ entries.length }}</span>
        <span v-if="hot > 0" class="dot" :title="t.board.needingAttention(hot)"></span>
      </button>
    </h3>

    <div v-if="!collapsed" class="group-list">
      <PrCard
        v-for="entry in entries"
        :key="entry.url"
        :entry="entry"
        :kind="kind"
        :show-url="showUrl"
        :show-repo="false"
        :now="now"
      />
    </div>
  </section>
</template>

<style scoped>
.group + .group { margin-top: 18px; }
.group-head { margin: 0 0 10px; }

/* A rule with a label on it rather than a raised block: the cards below are the
   panels, and a second frame around them only adds noise. */
.group-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 5px 8px;
  border: 1px solid transparent;
  border-bottom-color: var(--border);
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  background: transparent;
  color: var(--heading);
  font-size: 13px;
  font-weight: 600;
  text-align: left;
}
.group-toggle:hover:not(:disabled) {
  background: var(--bg-sunken);
  border-color: transparent;
  border-bottom-color: var(--border-strong);
}

.chev {
  flex: none;
  width: 11px;
  height: 11px;
  color: var(--text-faint);
  transition: transform var(--dur) var(--ease);
}
.chev-open { transform: rotate(90deg); }

.repo { font-size: 13px; }

/* The count and its dot are the tab strip's badges, so a repository heading
   and a tab report the same thing the same way. */
.count {
  min-width: 20px;
  padding: 0 6px;
  border-radius: var(--radius-full);
  background: var(--bg-sunken);
  color: var(--text-faint);
  font-size: 11px;
  font-weight: 600;
  text-align: center;
}
.count-hot { background: var(--dim-soft); color: var(--on-dim-soft); }
.dot {
  flex: none;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--dim);
}

/* One column on a laptop, more as the window allows. A card only needs about
   640px to hold its title and its indicators, and past that it was stretching
   to the far edge of a 4K display with everything still in the left third. */
.group-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 640px), 1fr));
  gap: 10px;
}
</style>
