<script setup lang="ts">
/**
 * The week's activity, as three small multiples over one shared scale.
 *
 * Why three strips rather than one stacked column per day: the app ships ten
 * themes, and a stacked bar would need three hues that stay tellable apart in
 * every one of them, including under colour-vision deficiency. Splitting the
 * series into its own row per metric removes the question entirely -- each
 * strip is a single series, so identity comes from the row's own label rather
 * than from a colour, and every bar can use the one hue the theme audit has
 * already cleared on this surface (--fact, >= 4:1 in all ten).
 *
 * The scale is shared across the three rows on purpose: it is what makes the
 * rows comparable, and reviewing five branches should look like more than
 * merging one.
 *
 * Hovering a day lights that column in all three rows and writes the day's
 * figures into the heading, rather than floating a box over the bars. Three
 * short rows leave a tooltip nowhere to go that is not on top of the data,
 * and a readout in the heading needs no positioning, never clips, and cannot
 * hang off the edge of the plot at the first or last column.
 */
import { computed, ref } from 'vue'
import { useI18n } from '../i18n'
import type { StatsDay } from '../lib/types'

const props = defineProps<{ days: StatsDay[] }>()

const { t, weekday, day, num } = useI18n()

/** Which day the pointer is on, lighting that column across all three rows. */
const hover = ref<number | null>(null)

type Row = { key: string; label: string; get: (d: StatsDay) => number }

const rows = computed<Row[]>(() => [
  { key: 'opened', label: t.value.dashboard.opened, get: (d) => d.opened },
  { key: 'merged', label: t.value.dashboard.merged, get: (d) => d.merged },
  { key: 'reviewed', label: t.value.dashboard.reviewed, get: (d) => d.reviewed },
])

/** The tallest single bar anywhere in the chart -- one scale for all rows. */
const max = computed(() =>
  Math.max(1, ...props.days.flatMap((d) => [d.opened, d.merged, d.reviewed])),
)

const empty = computed(() => props.days.every((d) => d.opened + d.merged + d.reviewed === 0))

/** True proportion of the track. A bar is never given a floor: a nudged-up
 *  height would misstate the value, and the track is tall enough that one
 *  out of nine still lands on a visible few pixels. */
function height(value: number): string {
  return `${(value / max.value) * 100}%`
}

const hovered = computed(() => (hover.value === null ? null : (props.days[hover.value] ?? null)))
</script>

<template>
  <figure class="chart">
    <figcaption class="chart-head">
      <div class="chart-titles">
        <h3>{{ t.dashboard.chartTitle }}</h3>
        <p class="soft">{{ t.dashboard.chartNote }}</p>
      </div>
      <!-- Always rendered, so the heading does not reflow on hover. -->
      <p class="readout" :class="{ 'readout-on': hovered }">
        <template v-if="hovered">
          <strong>{{ day(hovered.date) }}</strong>
          <span>{{ t.dashboard.opened }} {{ num(hovered.opened) }}</span>
          <span>{{ t.dashboard.merged }} {{ num(hovered.merged) }}</span>
          <span>{{ t.dashboard.reviewed }} {{ num(hovered.reviewed) }}</span>
        </template>
      </p>
    </figcaption>

    <p v-if="empty" class="chart-empty soft">{{ t.dashboard.chartEmpty }}</p>

    <div v-else class="plot" @mouseleave="hover = null">
      <template v-for="row in rows" :key="row.key">
        <span class="row-label">{{ row.label }}</span>
        <div class="track">
          <div
            v-for="(d, i) in days"
            :key="d.date"
            class="slot"
            :class="{ lit: hover === i }"
            @mouseenter="hover = i"
          >
            <span
              class="bar"
              :class="{ 'bar-zero': row.get(d) === 0 }"
              :style="row.get(d) > 0 ? { height: height(row.get(d)) } : undefined"
            ></span>
          </div>
        </div>
      </template>

      <span class="row-label"></span>
      <div class="axis">
        <span v-for="(d, i) in days" :key="d.date" class="tick" :class="{ lit: hover === i }">
          {{ weekday(d.date) }}
        </span>
      </div>
    </div>

    <!-- The same numbers as a table, for anything that cannot read the plot. -->
    <table class="sr-only">
      <caption>{{ t.dashboard.chartTable }}</caption>
      <thead>
        <tr>
          <th scope="col">{{ t.dashboard.columnDay }}</th>
          <th scope="col">{{ t.dashboard.opened }}</th>
          <th scope="col">{{ t.dashboard.merged }}</th>
          <th scope="col">{{ t.dashboard.reviewed }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in days" :key="d.date">
          <th scope="row">{{ day(d.date) }}</th>
          <td>{{ d.opened }}</td>
          <td>{{ d.merged }}</td>
          <td>{{ d.reviewed }}</td>
        </tr>
      </tbody>
    </table>
  </figure>
</template>

<style scoped>
.chart { margin: 0; flex: 1 1 340px; min-width: 0; }

.chart-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.chart-titles h3 { margin: 0; font-size: 14px; color: var(--heading); }
.chart-titles p { margin: 2px 0 0; font-size: 12.5px; }

/* The hover readout. It keeps its box whether or not a day is under the
   pointer, so lighting a column never nudges the heading around. */
.readout {
  margin: 0;
  min-height: 24px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 2px 10px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  font-size: 11.5px;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.readout-on { border-color: var(--border); background: var(--bg-sunken); }
.readout strong { color: var(--heading); font-weight: 600; }

.chart-empty { margin: 18px 0 0; }

.plot {
  margin-top: 16px;
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: end;
  gap: 10px 12px;
}

.row-label {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  padding-bottom: 2px;
}

.track {
  display: flex;
  align-items: flex-end;
  /* The surface gap: white is what separates neighbouring columns, not a
     stroke drawn around each one. */
  gap: 2px;
  height: 52px;
}

.slot {
  flex: 1 1 0;
  /* Bands are capped as well as bars. Seven days across a full-width card
     would leave each 24px bar marooned in a hundred pixels of air and stop
     reading as a chart at all, so the strip keeps its own width and lets the
     card's leftover go to the figures beside it. */
  max-width: 44px;
  height: 100%;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  transition: background var(--dur) var(--ease);
}
.slot.lit { background: var(--fact-soft); }

.bar {
  width: 100%;
  /* Capped, so a wide window leaves air in the band instead of filling it. */
  max-width: 24px;
  background: var(--fact);
  /* Rounded at the data end, square where it meets the baseline. */
  border-radius: 4px 4px 0 0;
}
/* A day with none of this metric: a flat rule on the baseline, so the column
   still reads as present rather than as missing data. */
.bar-zero { height: 2px; background: var(--border-strong); border-radius: 0; }

.axis { display: flex; gap: 2px; }
.tick {
  flex: 1 1 0;
  /* Matches the slot cap, so a tick stays under its own column. */
  max-width: 44px;
  text-align: center;
  font-size: 11px;
  color: var(--text-faint);
  white-space: nowrap;
}
.tick.lit { color: var(--on-fact-soft); font-weight: 600; }
</style>
