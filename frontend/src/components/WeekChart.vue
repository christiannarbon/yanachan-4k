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
 * Hovering a day lights that column in all three rows and names the figures in
 * a tooltip pinned above that column. The figures used to sit in the heading
 * instead, which put them beside the caption on a wide card and under it on a
 * narrow one -- the same hover landing in two different places. A band of its
 * own above the plot fixes where they appear: the box is measured and slid
 * back inside the plot at the first and last column, so it never hangs off
 * the card and never covers the bars it is describing.
 */
import { computed, nextTick, ref } from 'vue'
import { useI18n } from '../i18n'
import type { StatsDay } from '../lib/types'

const props = defineProps<{ days: StatsDay[] }>()

const { t, weekday, day, num } = useI18n()

/** Which day the pointer is on, lighting that column across all three rows. */
const hover = ref<number | null>(null)

const band = ref<HTMLElement | null>(null)
const tip = ref<HTMLElement | null>(null)

/** Everything below is measured in pixels from the left edge of the band. */
const column = ref(0)
const tipLeft = ref(0)
const arrowLeft = ref(0)
/** The box is hidden until it has been measured, so it never shows up at the
 *  left edge for a frame before sliding under its column. */
const placed = ref(false)

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

async function enter(i: number, event: MouseEvent) {
  hover.value = i
  const slot = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const box = band.value?.getBoundingClientRect()
  if (!box) return
  column.value = slot.left + slot.width / 2 - box.left
  await nextTick()
  place()
}

function leave() {
  hover.value = null
  placed.value = false
}

/** Centre the box on its column, then pull it back inside the band if that
 *  would push it off either end. The arrow stays on the column either way. */
function place() {
  const width = tip.value?.offsetWidth
  const room = band.value?.clientWidth
  if (!width || !room) return
  tipLeft.value = Math.max(0, Math.min(column.value - width / 2, room - width))
  arrowLeft.value = Math.min(Math.max(column.value - tipLeft.value, 12), width - 12)
  placed.value = true
}
</script>

<template>
  <figure class="chart">
    <figcaption class="chart-head">
      <h3>{{ t.dashboard.chartTitle }}</h3>
      <p class="soft">{{ t.dashboard.chartNote }}</p>
    </figcaption>

    <p v-if="empty" class="chart-empty soft">{{ t.dashboard.chartEmpty }}</p>

    <!-- The band is kept clear whether or not a day is hovered, so the tooltip
         has somewhere to sit that is neither on the bars nor on the caption. -->
    <div v-if="!empty" ref="band" class="tip-band">
      <div
        v-if="hovered"
        ref="tip"
        class="tip"
        :class="{ 'tip-on': placed }"
        :style="{ transform: `translateX(${tipLeft}px)` }"
      >
        <strong>{{ day(hovered.date) }}</strong>
        <span>{{ t.dashboard.opened }} {{ num(hovered.opened) }}</span>
        <span>{{ t.dashboard.merged }} {{ num(hovered.merged) }}</span>
        <span>{{ t.dashboard.reviewed }} {{ num(hovered.reviewed) }}</span>
        <span class="tip-arrow" :style="{ left: `${arrowLeft}px` }"></span>
      </div>
    </div>

    <div v-if="!empty" class="plot" @mouseleave="leave">
      <template v-for="row in rows" :key="row.key">
        <span class="row-label">{{ row.label }}</span>
        <div class="track">
          <div
            v-for="(d, i) in days"
            :key="d.date"
            class="slot"
            :class="{ lit: hover === i }"
            @mouseenter="enter(i, $event)"
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

.chart-head h3 { margin: 0; font-size: 14px; color: var(--heading); }
.chart-head p { margin: 2px 0 0; font-size: 12.5px; }

.chart-empty { margin: 18px 0 0; }

/* The tooltip's own strip of the card. Reserving it costs a line of height
   and buys a hover that always lands in the same place. */
.tip-band { position: relative; height: 34px; margin-top: 10px; }

/* Anchored by its bottom edge, so a box that wraps on a narrow card grows up
   into the caption rather than down over the first row of bars. Slid along by
   a transform rather than by `left`: an absolute box sizes itself against the
   room left of the band's right edge, so moving it with `left` would change
   the very width the move was measured from. */
.tip {
  position: absolute;
  left: 0;
  bottom: 8px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px 10px;
  max-width: 100%;
  /* The box sits above the columns it reads from; it must never take the
     pointer off one of them. */
  pointer-events: none;
  padding: 4px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-sunken);
  font-size: 11.5px;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
  opacity: 0;
}
.tip-on { opacity: 1; }
.tip span { white-space: nowrap; }
.tip strong { color: var(--heading); font-weight: 600; }

/* The nib: a square turned on its corner, with only the two lower edges
   stroked, so its fill covers the box's own border where the two meet. */
.tip-arrow {
  position: absolute;
  top: 100%;
  width: 7px;
  height: 7px;
  margin: -4px 0 0 -4px;
  border: 1px solid var(--border);
  border-top: 0;
  border-left: 0;
  background: var(--bg-sunken);
  transform: rotate(45deg);
}

.plot {
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
