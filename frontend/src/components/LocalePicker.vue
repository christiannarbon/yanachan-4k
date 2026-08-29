<script setup lang="ts">
/**
 * Language switch for the corner cluster.
 *
 * Two languages, so this is a segmented pair rather than a menu: both options
 * stay visible, each written in its own language, which is what someone
 * looking for their language actually scans for. The choice persists per
 * browser -- see i18n/index.ts.
 */
import { useI18n } from '../i18n'

const { locale, localeNames, t, setLocale } = useI18n()
</script>

<template>
  <div class="locales" role="group" :aria-label="t.locale.label">
    <button
      v-for="option in localeNames"
      :key="option.id"
      class="ghost opt"
      :class="{ 'opt-on': option.id === locale }"
      :aria-pressed="option.id === locale"
      @click="setLocale(option.id)"
    >
      {{ option.name }}
    </button>
  </div>
</template>

<style scoped>
.locales {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  background: var(--bg-sunken);
}

.opt {
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-size: 12.5px;
  line-height: 1.6;
  white-space: nowrap;
}

/* The selected language reads in the theme's fact colour, matching the
   "showing active only" toggle down in the tab row. */
.opt-on {
  background: var(--panel);
  border-color: var(--fact-line);
  color: var(--on-fact-soft);
  font-weight: 600;
}
.opt-on:hover:not(:disabled) { background: var(--panel); border-color: var(--fact-line); }
</style>
