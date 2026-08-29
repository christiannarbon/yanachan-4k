<script setup lang="ts">
/**
 * Language and theme, as one cluster for the page's top right corner.
 *
 * Neither of these acts on the board -- they are preferences you set once and
 * forget -- so they are kept out of the row that refreshes and filters, and
 * given the corner instead, where chrome is looked for.
 */
import LocalePicker from './LocalePicker.vue'
import ThemePicker from './ThemePicker.vue'
</script>

<template>
  <div class="corner">
    <LocalePicker />
    <ThemePicker />
  </div>
</template>

<style scoped>
.corner {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}
</style>
