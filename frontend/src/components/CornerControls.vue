<script setup lang="ts">
/**
 * Language, theme and settings, as one cluster for the page's top right corner.
 *
 * None of these acts on the board -- they are preferences you set once and
 * forget -- so they are kept out of the row that refreshes and filters, and
 * given the corner instead, where chrome is looked for.
 *
 * Settings is the odd one: it belongs here by the same reasoning, but there is
 * nothing to configure until you are signed in, so the sign-in screen leaves
 * the button out.
 */
import { useI18n } from '../i18n'
import LocalePicker from './LocalePicker.vue'
import ThemePicker from './ThemePicker.vue'

defineProps<{ withSettings?: boolean }>()
const emit = defineEmits<{ (e: 'settings'): void }>()

const { t } = useI18n()
</script>

<template>
  <div class="corner">
    <LocalePicker />
    <ThemePicker />
    <button
      v-if="withSettings"
      class="ghost gear"
      :title="t.settings.title"
      @click="emit('settings')"
    >
      <svg viewBox="0 0 16 16" aria-hidden="true">
        <circle cx="8" cy="8" r="2.4" fill="none" stroke="currentColor" stroke-width="1.4" />
        <path
          d="M8 1.2v1.7M8 13.1v1.7M1.2 8h1.7M13.1 8h1.7M3.2 3.2l1.2 1.2M11.6 11.6l1.2 1.2M12.8 3.2l-1.2 1.2M4.4 11.6l-1.2 1.2"
          fill="none"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
        />
      </svg>
      <span class="sr-only">{{ t.settings.title }}</span>
    </button>
  </div>
</template>

<style scoped>
.corner {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

/* Square, so it reads as an icon beside two worded controls rather than a
   button whose label failed to load. */
.gear {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-1);
  color: var(--text-muted);
}
.gear svg { width: 16px; height: 16px; }
</style>
