<script setup lang="ts">
import { useI18n } from '../i18n'
import type { Section } from '../lib/types'

defineProps<{ sections: Section[]; activeId: string }>()
const emit = defineEmits<{ (e: 'select', id: string): void }>()

const { t, sectionTitle } = useI18n()
</script>

<template>
  <nav class="tabs" role="tablist">
    <button
      role="tab"
      class="tab tab-dashboard"
      :class="{ 'tab-active': activeId === 'dashboard' }"
      :aria-selected="activeId === 'dashboard'"
      @click="emit('select', 'dashboard')"
    >
      {{ t.sections.dashboard }}
    </button>
    <button
      v-for="section in sections"
      :key="section.id"
      role="tab"
      class="tab"
      :class="{ 'tab-active': section.id === activeId }"
      :aria-selected="section.id === activeId"
      @click="emit('select', section.id)"
    >
      <span class="tab-label">{{ sectionTitle(section) }}</span>
      <span class="tab-count" :class="{ 'tab-count-hot': section.hot > 0 }">{{ section.total }}</span>
      <span v-if="section.hot > 0" class="tab-dot" :title="t.board.needingAttention(section.hot)"></span>
    </button>
    <span class="spacer"></span>
    <button
      role="tab"
      class="tab tab-settings"
      :class="{ 'tab-active': activeId === 'settings' }"
      :aria-selected="activeId === 'settings'"
      @click="emit('select', 'settings')"
    >
      {{ t.sections.settings }}
    </button>
  </nav>
</template>

<style scoped>
.tabs {
  display: flex;
  align-items: stretch;
  gap: 2px;
  border-bottom: 1px solid var(--border);
  overflow-x: auto;
  /* overflow-x alone also makes overflow-y scrollable, which shows a stray
     scrollbar because the active tab overhangs the strip by 1px. */
  overflow-y: hidden;
  scrollbar-width: thin;
}
.tabs::-webkit-scrollbar { height: 6px; }
.tabs::-webkit-scrollbar-thumb { background: var(--border-strong); border-radius: 3px; }
.tab {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 9px 14px;
  border: 1px solid transparent;
  border-bottom: none;
  border-radius: var(--radius) var(--radius) 0 0;
  background: transparent;
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  position: relative;
  bottom: -1px;
}
.tab:hover:not(.tab-active) { background: var(--fact-soft); border-color: transparent; }
.tab-active {
  background: var(--panel-raised);
  border-color: var(--border);
  border-bottom: 1px solid var(--panel-raised);
  color: var(--heading);
  font-weight: 600;
}
.tab-label { max-width: 260px; overflow: hidden; text-overflow: ellipsis; }
.tab-count {
  min-width: 20px;
  padding: 0 6px;
  border-radius: 999px;
  background: var(--bg-sunken);
  color: var(--text-faint);
  font-size: 11px;
  font-weight: 600;
  text-align: center;
}
.tab-count-hot { background: var(--dim-soft); color: var(--on-dim-soft); }
.tab-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--dim);
}
.tab-settings { margin-left: auto; }
/* The landing tab is a heading, not a count, so it carries no badge; the rule
   keeps it from sitting flush against the first section tab. */
.tab-dashboard { margin-right: 6px; }
</style>
