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
    <button v-if="withSettings" class="ghost open" @click="emit('settings')">
      {{ t.settings.title }}
    </button>
  </div>
</template>

<style scoped>
.corner {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

/* Worded rather than a cog, because both of its neighbours are worded and a
   16px gear reads as a flower. */
.open { font-size: 13px; white-space: nowrap; }
</style>
