<script setup lang="ts">
/**
 * A modal dialog: a titled panel over the page, with a close button.
 *
 * Built on the browser's own <dialog>, so the focus trap, the page going inert
 * behind it and Escape are the platform's rather than ours. Mounting opens it,
 * which means the caller decides whether it exists at all with v-if and gets a
 * fresh panel each time -- what it holds never has to remember being closed.
 */
import { onMounted, ref, useId } from 'vue'

import { useI18n } from '../i18n'

defineProps<{ title: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const { t } = useI18n()
const el = ref<HTMLDialogElement | null>(null)
const titleId = useId()

onMounted(() => el.value?.showModal())

/** A click that lands on the dialog itself came from the backdrop, because the
 *  panel inside covers every other pixel of it. */
function onClick(e: MouseEvent) {
  if (e.target === el.value) emit('close')
}
</script>

<template>
  <dialog
    ref="el"
    class="modal"
    :aria-labelledby="titleId"
    @cancel.prevent="emit('close')"
    @click="onClick"
  >
    <!-- Focus starts on the panel rather than on the first control, which
         would otherwise be Close, wearing a focus ring and looking pressed. -->
    <div class="panel" tabindex="-1" autofocus>
      <header class="head">
        <h2 :id="titleId">{{ title }}</h2>
        <button class="ghost" @click="emit('close')">{{ t.common.close }}</button>
      </header>
      <div class="body">
        <slot />
      </div>
    </div>
  </dialog>
</template>

<style scoped>
/* The dialog element is only the frame: it carries no colour of its own, and
   the panel inside is what you see. */
.modal {
  width: min(680px, calc(100vw - var(--space-8)));
  max-width: none;
  max-height: calc(100vh - var(--space-16));
  padding: 0;
  border: 0;
  background: transparent;
  overflow: visible;
}
/* A backdrop does not always inherit the page's custom properties, so the
   theme's overlay is asked for with a literal to fall back on. */
.modal::backdrop { background: var(--overlay, rgba(0, 0, 0, 0.42)); }

.panel {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - var(--space-16));
  background: var(--panel-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}
.panel:focus { outline: none; }

/* The title stays put while the body scrolls under it, so a long panel never
   leaves you scrolling to find out what you have open. */
.head {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--border);
}
.head h2 { margin: 0; font-size: 15px; color: var(--heading); }

.body {
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0 var(--space-5);
}
</style>
