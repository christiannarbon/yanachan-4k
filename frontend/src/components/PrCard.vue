<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '../i18n'
import type { Entry } from '../lib/types'

const props = withDefaults(
  defineProps<{
    entry: Entry
    showUrl: boolean
    now: number
    kind: string
    /** Off inside a repository group, whose heading already names the repo. */
    showRepo?: boolean
  }>(),
  { showRepo: true },
)

const { t, ago, actors } = useI18n()

const statusLabel = computed(() => {
  switch (props.entry.status) {
    case 'reply':
      return t.value.pr.statusReply
    case 'new':
      return t.value.pr.statusNew
    default:
      return ''
  }
})

const checksLabel = computed(() => {
  switch (props.entry.checks) {
    case 'success':
      return t.value.pr.checksSuccess
    case 'failure':
      return t.value.pr.checksFailure
    case 'pending':
      return t.value.pr.checksPending
    default:
      return ''
  }
})

const decisionLabel = computed(() => {
  switch (props.entry.reviewDecision) {
    case 'approved':
      return t.value.pr.approved
    case 'changes_requested':
      return t.value.pr.changesRequested
    case 'review_required':
      return t.value.pr.reviewRequired
    default:
      return ''
  }
})

const yourStateLabel = computed(() => {
  switch (props.entry.yourState) {
    case 'approved':
      return t.value.pr.youApproved
    case 'changes_requested':
      return t.value.pr.youRequestedChanges
    case 'commented':
      return t.value.pr.youCommented
    default:
      return ''
  }
})

const activitySummary = computed(() => {
  const e = props.entry
  const pr = t.value.pr
  const parts: string[] = []
  if (e.status === 'reply') {
    if (e.humanActors.length) parts.push(pr.answered(actors(e.humanActors)))
    if (e.botActors.length) parts.push(pr.bots(actors(e.botActors)))
  } else if (e.status === 'new') {
    if (props.kind === 'mine') {
      if (e.humanCount) parts.push(pr.commentsFrom(e.humanCount, actors(e.humanActors)))
      if (e.botCount) parts.push(pr.botUpdates(e.botCount))
    } else {
      if (e.humanActors.length) parts.push(actors(e.humanActors))
      if (e.botActors.length) parts.push(pr.bots(actors(e.botActors)))
    }
  }
  return parts.join(pr.partSeparator)
})

const quietLabel = computed(() => {
  const e = props.entry
  if (e.status !== 'quiet') return ''
  if (props.kind === 'mine') {
    return e.lastActivityAt
      ? t.value.pr.quietSince(ago(e.lastActivityAt, props.now))
      : t.value.pr.noComments
  }
  return t.value.pr.noActivityInWindow
})

const awaitingLabel = computed(() => {
  if (props.entry.awaiting === 'you') return t.value.pr.awaitingYou
  if (props.entry.awaiting === 'team') return t.value.pr.awaitingTeam
  return ''
})

const tone = computed(() => (props.entry.hot ? 'hot' : props.entry.active ? 'active' : 'quiet'))
</script>

<template>
  <article class="pr" :class="`pr-${tone}`">
    <div class="pr-head">
      <span class="badge" :class="entry.status === 'reply' ? 'badge-orange' : 'badge-green'" v-if="statusLabel">
        {{ statusLabel }}
      </span>
      <span class="repo mono" v-if="showRepo">{{ entry.repo }}</span>
      <span class="num mono">#{{ entry.number }}</span>
      <a class="title" :href="entry.url" target="_blank" rel="noopener noreferrer">{{ entry.title }}</a>
      <span class="badge badge-neutral" v-if="entry.isDraft">{{ t.pr.draft }}</span>
    </div>

    <div class="pr-meta row wrap">
      <span class="badge badge-orange" v-if="awaitingLabel">{{ awaitingLabel }}</span>
      <span class="badge badge-outline" v-if="yourStateLabel">
        {{ yourStateLabel }}<template v-if="entry.yourLastAt"> · {{ ago(entry.yourLastAt, now) }}</template>
      </span>
      <span class="badge badge-outline" v-if="entry.alsoRequestedFromYou">{{ t.pr.alsoRequestedFromYou }}</span>
      <span
        class="badge"
        v-if="decisionLabel"
        :class="entry.reviewDecision === 'approved' ? 'badge-green' : 'badge-orange'"
      >{{ decisionLabel }}</span>
      <span
        class="badge"
        v-if="checksLabel"
        :class="entry.checks === 'success' ? 'badge-green' : entry.checks === 'failure' ? 'badge-orange' : 'badge-neutral'"
      >{{ checksLabel }}</span>
    </div>

    <p class="pr-line">
      <span v-if="activitySummary" class="activity">{{ activitySummary }}</span>
      <span v-if="activitySummary && entry.lastActivityAt" class="muted"> · {{ ago(entry.lastActivityAt, now) }}</span>
      <span v-if="quietLabel" class="muted">{{ quietLabel }}</span>
      <span class="muted">{{ t.pr.byline(entry.author, ago(entry.createdAt, now)) }}</span>
    </p>

    <a v-if="showUrl" class="pr-url mono" :href="entry.url" target="_blank" rel="noopener noreferrer">{{ entry.url }}</a>
  </article>
</template>

<style scoped>
.pr {
  padding: 12px 16px 12px 14px;
  border: 1px solid var(--border);
  border-left-width: 4px;
  border-radius: var(--radius);
  background: var(--panel-raised);
  box-shadow: var(--shadow);
}
.pr-hot { border-left-color: var(--dim); }
.pr-active { border-left-color: var(--fact); }
.pr-quiet { border-left-color: var(--border-strong); }

.pr-head {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8px;
}
.repo { color: var(--on-fact-soft); font-size: 12.5px; }
.num { color: var(--text-faint); font-size: 12.5px; }
.title {
  font-weight: 600;
  color: var(--text);
  text-decoration: none;
  flex: 1 1 320px;
  min-width: 0;
}
.title:hover { color: var(--on-fact-soft); text-decoration: underline; }

.pr-meta { margin-top: 7px; gap: 6px; }
.pr-meta:empty { display: none; }

.pr-line { margin: 7px 0 0; font-size: 13px; color: var(--text-muted); }
.activity { color: var(--on-fact-soft); font-weight: 500; }

.pr-url {
  display: inline-block;
  margin-top: 6px;
  font-size: 11.5px;
  color: var(--text-faint);
  text-decoration: none;
  word-break: break-all;
}
.pr-url:hover { color: var(--accent); text-decoration: underline; }
</style>
