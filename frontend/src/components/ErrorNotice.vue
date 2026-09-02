<script setup lang="ts">
/**
 * A refresh that did not come back, said calmly.
 *
 * GitHub's edge answers a 502 from time to time, and the whole of it -- a page
 * of nginx HTML -- used to be the entire dashboard. The backend rides those out
 * now and words what is left as a sentence, so all this has to do is put the
 * plain part first and offer the one thing worth doing about it.
 *
 * The detail line is kept: when the trouble is a real outage, GitHub's own
 * wording is the useful half.
 */

import { useI18n } from '../i18n'

defineProps<{
  /** GitHub's own wording, or whatever the API sent back. */
  message: string
  /** A retry already in flight. */
  busy?: boolean
}>()

const emit = defineEmits<{ (e: 'retry'): void }>()

const { t } = useI18n()
</script>

<template>
  <div class="notice trouble">
    <div class="trouble-text">
      <p class="trouble-head">{{ t.board.troubleTitle }}</p>
      <p class="trouble-detail">{{ message }}</p>
    </div>
    <button class="ghost trouble-retry" :disabled="busy" @click="emit('retry')">
      {{ busy ? t.board.refreshing : t.board.tryAgain }}
    </button>
  </div>
</template>

<style scoped>
.trouble {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.trouble-text { min-width: 0; }
.trouble-head { margin: 0; font-weight: 600; }
/* Long, and occasionally a URL with no spaces in it. */
.trouble-detail { margin: 2px 0 0; font-size: 12.5px; overflow-wrap: anywhere; }
.trouble-retry {
  flex: none;
  border-color: var(--dim-line);
  color: var(--on-dim-soft);
}
.trouble-retry:hover:not(:disabled) { background: var(--panel); border-color: var(--dim-line); }
</style>
