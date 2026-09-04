<script setup lang="ts">
/**
 * Theme picker for the corner cluster.
 *
 * A plain button that opens the list of available palettes, grouped by kind:
 * Yanami's calorie meter first, then the app's own, then the painting-derived
 * ones. There is no light/dark control beside it -- the app is light-mode only,
 * so a theme is one palette.
 *
 * The palettes themselves live in styles/art-themes.css.
 */
import { computed, onBeforeUnmount, ref, watch, nextTick } from 'vue'

import { useTheme, type ArtTheme } from '../composables/useTheme'
import { useI18n } from '../i18n'

const { art, themes, current, setArt } = useTheme()
const { t } = useI18n()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const list = ref<HTMLElement | null>(null)

const label = computed(() =>
  current.value.subtitle ? `${current.value.name} — ${current.value.subtitle}` : current.value.name,
)

/** Group headings. The theme names themselves are proper nouns, so they stand
 *  as the generator wrote them in every language. */
const groups = computed<Record<ArtTheme['kind'], string>>(() => ({
  ode: t.value.theme.groupOde,
  house: t.value.theme.groupHouse,
  painting: t.value.theme.groupPainting,
}))

/**
 * Heading to print above each entry, keyed by index -- set only on the first
 * theme of each kind, so the list reads as groups without nesting the markup.
 * The themes arrive already sorted by kind from the generator.
 */
const headings = computed(() => {
  const seen = new Set<string>()
  return themes.map((theme) => {
    if (seen.has(theme.kind)) return null
    seen.add(theme.kind)
    return groups.value[theme.kind]
  })
})

/**
 * The menu closes first, and only then does the theme change.
 *
 * The sweep snapshots the page as it stands, so setting the theme while the
 * menu is still on screen would drag an open list across the window and drop
 * it at the far edge. A tick is enough for the list to be gone from both
 * frames, and is not long enough to feel like a delay.
 */
async function choose(id: string) {
  open.value = false
  await nextTick()
  setArt(id)
}

function onDocPointer(e: PointerEvent) {
  if (!root.value?.contains(e.target as Node)) open.value = false
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    e.stopPropagation()
    open.value = false
  }
}

watch(open, async (v) => {
  if (v) {
    document.addEventListener('pointerdown', onDocPointer)
    await nextTick()
    list.value?.querySelector<HTMLElement>('[data-selected="true"]')?.focus()
  } else {
    document.removeEventListener('pointerdown', onDocPointer)
  }
})

onBeforeUnmount(() => document.removeEventListener('pointerdown', onDocPointer))
</script>

<template>
  <div ref="root" class="picker" @keydown="onKey">
    <button
      class="ghost trigger"
      :title="t.theme.current(label)"
      aria-haspopup="listbox"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span class="dot" :style="{ background: current.swatch }" aria-hidden="true" />
      <span class="name">{{ current.name }}</span>
      <span class="caret" aria-hidden="true">▾</span>
    </button>

    <div v-if="open" ref="list" class="menu" role="listbox" :aria-label="t.theme.label">
      <template v-for="(theme, i) in themes" :key="theme.id">
        <p v-if="headings[i]" class="menu-head" :class="{ 'menu-head--mid': i > 0 }">
          {{ headings[i] }}
        </p>
        <button
          class="opt"
          role="option"
          :aria-selected="theme.id === art"
          :data-selected="theme.id === art"
          @click="choose(theme.id)"
        >
          <span class="dot" :style="{ background: theme.swatch }" aria-hidden="true" />
          <span class="opt-text">
            <span class="opt-name">{{ theme.name }}</span>
            <span class="opt-sub">{{ theme.subtitle }}</span>
          </span>
          <span v-if="theme.id === art" class="tick" aria-hidden="true">✓</span>
        </button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.picker { position: relative; }

.trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 13px;
}
.name {
  max-width: 12ch;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.caret { font-size: 9px; color: var(--text-faint); }

.dot {
  width: 10px;
  height: 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-strong);
  flex: none;
}

.menu {
  position: absolute;
  top: calc(100% + var(--space-2));
  right: 0;
  z-index: 40;
  width: 264px;
  padding: var(--space-1);
  background: var(--panel-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
}

.menu-head {
  margin: 0;
  padding: var(--space-2) var(--space-3) var(--space-1);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  color: var(--text-faint);
}
.menu-head--mid {
  margin-top: var(--space-1);
  padding-top: var(--space-3);
  border-top: 1px solid var(--border);
}

.opt {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: none;
  border: 0;
  border-radius: var(--radius);
  text-align: left;
  transition: background var(--dur) var(--ease);
}
.opt:hover { background: var(--bg-sunken); border-color: transparent; }
.opt[data-selected="true"] { background: var(--fact-soft); }

.opt-text { display: flex; flex-direction: column; min-width: 0; flex: 1; }
.opt-name {
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}
.opt-sub {
  font-size: 11px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tick { color: var(--accent); font-size: 12px; }
</style>
