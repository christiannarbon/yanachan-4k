<script setup lang="ts">
/**
 * The landing tab: the viewer's own week.
 *
 * Every other tab in this app answers "what needs you". This one answers
 * "what did you get done", which is a different job and gets a different
 * shape -- figures rather than a list, and copy that never counts what did
 * not happen. A week with nothing in it gets one sentence, not four zeroes.
 *
 * The calorie total is the hero figure, and the only one on the page. It is
 * the joke the repository is named after: 4K is a calorie count, and under
 * the Yanami theme calorie-meter.css sets it in the kcal tiles the official
 * site uses for exactly this.
 */
import { computed } from 'vue'
import WeekChart from './WeekChart.vue'
import { useI18n } from '../i18n'
import type { Stats } from '../lib/types'

const props = defineProps<{ stats: Stats }>()

const { t, day, dayOf, num } = useI18n()

const s = computed(() => props.stats)

/** Nothing opened, merged, closed or reviewed: the week gets a sentence. */
const quiet = computed(
  () => s.value.opened + s.value.merged + s.value.closed + s.value.reviewed === 0,
)

const range = computed(() => t.value.dashboard.range(dayOf(s.value.week.since), dayOf(s.value.week.until)))

const tiles = computed(() => [
  { key: 'opened', label: t.value.dashboard.opened, value: s.value.opened, note: t.value.dashboard.openedNote },
  { key: 'merged', label: t.value.dashboard.merged, value: s.value.merged, note: t.value.dashboard.mergedNote },
  { key: 'closed', label: t.value.dashboard.closed, value: s.value.closed, note: t.value.dashboard.closedNote },
  { key: 'reviewed', label: t.value.dashboard.reviewed, value: s.value.reviewed, note: t.value.dashboard.reviewedNote },
])

/** The line under the chart. Each figure is dropped when it is zero, so the
 *  row reads as a list of things done rather than a form with blanks. */
const totals = computed(() => {
  const d = t.value.dashboard
  const out: string[] = []
  if (s.value.reviewsWritten) out.push(d.reviewsWritten(s.value.reviewsWritten))
  if (s.value.approvals) out.push(d.approvalsGiven(s.value.approvals))
  if (s.value.repos) out.push(d.reposTouched(s.value.repos))
  if (s.value.additions || s.value.deletions) {
    out.push(d.linesMerged(num(s.value.additions), num(s.value.deletions)))
  }
  if (s.value.filesChanged) out.push(d.filesChanged(s.value.filesChanged))
  return out
})

/** The busiest day is read off the chart's own data rather than sent down. */
const busiestDay = computed(() => {
  let best: (typeof s.value.daily)[number] | null = null
  for (const d of s.value.daily) {
    const total = d.opened + d.merged + d.reviewed
    if (total > 0 && (!best || total > best.opened + best.merged + best.reviewed)) best = d
  }
  return best
})
</script>

<template>
  <div class="stats">
    <section v-if="quiet" class="card block quiet">
      <h2>{{ t.dashboard.quietTitle }}</h2>
      <p class="soft">{{ t.dashboard.quietBody }}</p>
      <p class="range soft">{{ range }}</p>
    </section>

    <template v-else>
      <section class="card block">
        <p class="hero">
          <span class="hero-num" :title="t.dashboard.kcalHow">{{ num(s.kcal) }}</span>
          <span class="hero-unit">{{ t.dashboard.kcalUnit }}</span>
        </p>
        <p class="hero-caption soft">
          {{ t.dashboard.kcalCaption }} · {{ range }}
          <template v-if="s.activeDays"> · {{ t.dashboard.activeDays(s.activeDays, s.week.days) }}</template>
          <template v-if="s.streak"> · <span class="streak">{{ t.dashboard.streak(s.streak) }}</span></template>
        </p>

        <ul class="tiles">
          <li v-for="tile in tiles" :key="tile.key" class="tile">
            <span class="tile-label">{{ tile.label }}</span>
            <span class="tile-value">{{ num(tile.value) }}</span>
            <span class="tile-note soft">{{ tile.note }}</span>
          </li>
        </ul>
      </section>

      <section class="card block chart-card">
        <WeekChart :days="s.daily" />
        <ul v-if="totals.length" class="totals soft">
          <li v-for="(fact, i) in totals" :key="i">{{ fact }}</li>
        </ul>
      </section>

      <section
        v-if="s.highlights.fastestMerge || s.highlights.biggestMerge || s.highlights.topRepo || busiestDay"
        class="card block"
      >
        <h2>{{ t.dashboard.highlightsTitle }}</h2>
        <dl class="highlights">
          <template v-if="s.highlights.fastestMerge">
            <dt>{{ t.dashboard.fastestMerge }}</dt>
            <dd>
              <a :href="s.highlights.fastestMerge.url" target="_blank" rel="noopener noreferrer">
                <span class="mono">{{ s.highlights.fastestMerge.repo }} #{{ s.highlights.fastestMerge.number }}</span>
              </a>
              <span class="soft"> · {{ t.dashboard.duration(s.highlights.fastestMinutes) }}</span>
            </dd>
          </template>

          <template v-if="s.highlights.biggestMerge">
            <dt>{{ t.dashboard.biggestMerge }}</dt>
            <dd>
              <a :href="s.highlights.biggestMerge.url" target="_blank" rel="noopener noreferrer">
                <span class="mono">{{ s.highlights.biggestMerge.repo }} #{{ s.highlights.biggestMerge.number }}</span>
              </a>
              <span class="soft"> · {{ t.dashboard.lineCount(num(s.highlights.biggestLines)) }}</span>
            </dd>
          </template>

          <template v-if="s.highlights.topRepo">
            <dt>{{ t.dashboard.busiestRepo }}</dt>
            <dd>
              <span class="mono">{{ s.highlights.topRepo }}</span>
              <span class="soft"> · {{ t.dashboard.prCount(s.highlights.topRepoCount) }}</span>
            </dd>
          </template>

          <template v-if="busiestDay">
            <dt>{{ t.dashboard.busiestDay }}</dt>
            <dd>
              {{ day(busiestDay.date) }}
              <span class="soft">
                · {{ t.dashboard.prCount(busiestDay.opened + busiestDay.merged + busiestDay.reviewed) }}
              </span>
            </dd>
          </template>
        </dl>
      </section>
    </template>
  </div>
</template>

<style scoped>
.stats { display: flex; flex-direction: column; gap: 14px; padding: 18px 0 40px; }
.block { padding: 18px 20px; }
.block h2 { margin: 0 0 8px; font-size: 15px; color: var(--heading); }

.quiet .range { margin: 12px 0 0; font-size: 12.5px; }
.quiet p { margin: 0; }

/* The hero figure. Proportional digits, not tabular: at this size tabular
   gives every digit the width of a zero and the number reads loose. */
.hero { margin: 0; display: flex; align-items: baseline; gap: 8px; }
.hero-num {
  font-size: 52px;
  line-height: 1.05;
  font-weight: 700;
  color: var(--heading);
  letter-spacing: -0.02em;
}
.hero-unit { font-size: 17px; font-weight: 600; color: var(--text-muted); }
.hero-caption { margin: 4px 0 0; font-size: 12.5px; }
.streak { color: var(--on-dim-soft); font-weight: 600; }

.tiles {
  list-style: none;
  margin: 18px 0 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(132px, 1fr));
  gap: 10px;
}
.tile {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-sunken);
}
.tile-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-muted);
}
html[lang='ja'] .tile-label { letter-spacing: 0; }
.tile-value { font-size: 26px; font-weight: 700; line-height: 1.15; color: var(--heading); }
.tile-note { font-size: 11.5px; }

/* The chart keeps its own width, so the week's other figures take the rest of
   the card rather than leaving it empty. Both are flex items: below the
   breakpoint the list simply wraps under the strip. */
.chart-card {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 20px 36px;
}
.totals {
  flex: 1 1 200px;
  align-self: center;
  list-style: none;
  margin: 0;
  padding: 0;
  font-size: 12.5px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.totals li { padding-left: 12px; position: relative; }
.totals li::before { content: '·'; position: absolute; left: 0; color: var(--text-faint); }

.highlights {
  margin: 0;
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 6px 16px;
  font-size: 13px;
}
.highlights dt { color: var(--text-muted); white-space: nowrap; }
.highlights dd { margin: 0; min-width: 0; }
.highlights a { text-decoration: none; }
.highlights a:hover { text-decoration: underline; }
</style>
