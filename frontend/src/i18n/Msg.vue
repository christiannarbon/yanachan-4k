<script setup lang="ts">
/**
 * Renders a catalog string that carries inline markup.
 *
 * A sentence like "Open {link} and enter this code:" cannot be assembled from
 * fragments, because Japanese puts the marked-up piece somewhere English does
 * not. So the whole sentence stays one translatable string with {name}
 * placeholders, and each placeholder is filled by the same-named slot:
 *
 *   <Msg :text="t.auth.oauthOpen">
 *     <template #link><a :href="uri">{{ uri }}</a></template>
 *   </Msg>
 *
 * The slot content is real markup rather than interpolated HTML, so a repo
 * name or a filesystem path passing through is never parsed as a tag.
 */
import { computed } from 'vue'

const props = defineProps<{ text: string }>()

const PLACEHOLDER = /\{[a-zA-Z][a-zA-Z0-9]*\}/

/** Splits "Open {link} and…" into a literal, a slot, and another literal. */
const parts = computed(() =>
  props.text
    .split(new RegExp(`(${PLACEHOLDER.source})`))
    .filter((p) => p !== '')
    .map((p) =>
      new RegExp(`^${PLACEHOLDER.source}$`).test(p)
        ? { slot: p.slice(1, -1), text: '' }
        : { slot: '', text: p },
    ),
)
</script>

<template><template v-for="(part, i) in parts" :key="i"><slot v-if="part.slot" :name="part.slot" /><template v-else>{{ part.text }}</template></template></template>
