<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import AuthGate from './components/AuthGate.vue'
import LocalePicker from './components/LocalePicker.vue'
import PrCard from './components/PrCard.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import TabBar from './components/TabBar.vue'
import ThemePicker from './components/ThemePicker.vue'
import Msg from './i18n/Msg.vue'
import { useI18n } from './i18n'
import { ApiError, api } from './lib/api'
import type { AuthStatus, Board, Settings } from './lib/types'

const { t, ago, stamp, sectionTitle, windowLabel } = useI18n()

const authStatus = ref<AuthStatus | null>(null)
const board = ref<Board | null>(null)
const settings = ref<Settings | null>(null)
const activeTab = ref('mine')
const loading = ref(true)
const refreshing = ref(false)
const error = ref('')
const autoRefresh = ref(false)
const now = ref(Date.now())

let clockTimer: number | undefined
let refreshTimer: number | undefined
const REFRESH_MS = 5 * 60 * 1000

onMounted(async () => {
  clockTimer = window.setInterval(() => (now.value = Date.now()), 30_000)
  await bootstrap()
})

onUnmounted(() => {
  if (clockTimer !== undefined) window.clearInterval(clockTimer)
  stopAutoRefresh()
})

watch(autoRefresh, (on) => (on ? startAutoRefresh() : stopAutoRefresh()))

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = window.setInterval(() => void loadBoard(), REFRESH_MS)
}
function stopAutoRefresh() {
  if (refreshTimer !== undefined) window.clearInterval(refreshTimer)
  refreshTimer = undefined
}

async function bootstrap() {
  loading.value = true
  error.value = ''
  try {
    authStatus.value = await api.authStatus()
    if (authStatus.value.authenticated) {
      settings.value = await api.settings()
      await loadBoard()
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function loadBoard() {
  refreshing.value = true
  error.value = ''
  try {
    const next = await api.board()
    board.value = next
    now.value = Date.now()
    if (!next.sections.some((s) => s.id === activeTab.value) && activeTab.value !== 'settings') {
      activeTab.value = next.sections[0]?.id ?? 'mine'
    }
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      board.value = null
      authStatus.value = await api.authStatus()
    } else {
      error.value = (e as Error).message
    }
  } finally {
    refreshing.value = false
  }
}

async function onAuthenticated() {
  await bootstrap()
}

async function onSignOut() {
  await api.logout()
  board.value = null
  settings.value = null
  activeTab.value = 'mine'
  authStatus.value = await api.authStatus()
}

async function onSettingsSaved(saved: Settings) {
  settings.value = saved
  await loadBoard()
}

const currentSection = computed(() => board.value?.sections.find((s) => s.id === activeTab.value) ?? null)
const showUrls = computed(() => settings.value?.showUrls ?? true)
const totalHot = computed(() => board.value?.sections.reduce((sum, s) => sum + s.hot, 0) ?? 0)

async function toggleOnlyActive() {
  if (!settings.value) return
  const next = { ...settings.value, onlyActive: !settings.value.onlyActive }
  settings.value = await api.saveSettings(next)
  await loadBoard()
}
</script>

<template>
  <div class="shell">
    <p v-if="loading" class="centered soft">{{ t.common.loading }}</p>

    <AuthGate
      v-else-if="authStatus && !authStatus.authenticated"
      :status="authStatus"
      @authenticated="onAuthenticated"
    />

    <template v-else-if="board && settings">
      <header class="topbar">
        <div class="topbar-inner">
          <div class="brand">
            <h1>{{ t.appName }}</h1>
            <p class="brand-sub soft">
              <Msg :text="t.board.summary">
                <template #login><span class="mono">@{{ board.login }}</span></template>
                <template #label>{{ windowLabel(board.window) }}</template>
                <template #stamp>{{ stamp(board.window.cutoff) }}</template>
              </Msg>
              <template v-if="totalHot > 0">
                ·
                <span class="attention">{{ t.board.needingAttention(totalHot) }}</span>
              </template>
            </p>
          </div>
          <div class="row controls">
            <button class="ghost" :class="{ toggled: settings.onlyActive }" @click="toggleOnlyActive">
              {{ settings.onlyActive ? t.board.showingActiveOnly : t.board.showingEverything }}
            </button>
            <label class="auto soft">
              <input type="checkbox" v-model="autoRefresh" />
              {{ t.board.autoRefresh }}
            </label>
            <button class="primary" :disabled="refreshing" @click="loadBoard">
              {{ refreshing ? t.board.refreshing : t.board.refresh }}
            </button>
            <LocalePicker />
            <ThemePicker />
          </div>
        </div>
        <div class="topbar-inner tabs-row">
          <TabBar :sections="board.sections" :active-id="activeTab" @select="activeTab = $event" />
        </div>
      </header>

      <main class="content">
        <p v-if="error" class="notice">{{ error }}</p>
        <p v-if="board.warning" class="notice">{{ t.board.githubReported(board.warning) }}</p>

        <SettingsPanel
          v-if="activeTab === 'settings'"
          :settings="settings"
          :session-mode="board.authMode"
          :login="board.login"
          @saved="onSettingsSaved"
          @signout="onSignOut"
        />

        <template v-else-if="currentSection">
          <div class="section-head">
            <h2>{{ sectionTitle(currentSection) }}</h2>
            <p class="soft">
              {{ t.board.prCount(currentSection.total) }}
              <template v-if="currentSection.hot > 0">
                · {{ t.board.needingAttention(currentSection.hot) }}
              </template>
              <template v-if="settings.onlyActive && currentSection.total !== currentSection.entries.length">
                · {{ t.board.hiddenAsQuiet(currentSection.total - currentSection.entries.length) }}
              </template>
            </p>
          </div>

          <p v-if="currentSection.error" class="notice">{{ currentSection.error }}</p>

          <div v-if="currentSection.entries.length" class="list">
            <PrCard
              v-for="entry in currentSection.entries"
              :key="entry.url"
              :entry="entry"
              :kind="currentSection.kind"
              :show-url="showUrls"
              :now="now"
            />
          </div>
          <p v-else class="card empty-state soft">
            {{ t.board.empty }}
            <Msg
              v-if="currentSection.kind === 'team' || currentSection.kind === 'org'"
              :text="t.board.emptyScope"
            >
              <template #ref><span class="mono">{{ currentSection.ref }}</span></template>
            </Msg>
          </p>
        </template>

        <footer class="foot soft">
          {{ t.board.footer(ago(board.generatedAt, now), board.limit) }}
        </footer>
      </main>
    </template>

    <p v-else class="centered notice">{{ error || t.common.somethingWrong }}</p>
  </div>
</template>

<style scoped>
.shell { min-height: 100%; display: flex; flex-direction: column; }
.centered { padding: 60px 20px; text-align: center; }

.topbar {
  background: var(--panel);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 10;
}
.topbar-inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.topbar-inner:first-child { padding-top: 16px; padding-bottom: 12px; }
.tabs-row { display: block; padding-bottom: 0; }

.brand h1 {
  margin: 0;
  font-size: 17px;
  letter-spacing: -0.01em;
  color: var(--heading);
}
.brand-sub { margin: 2px 0 0; font-size: 12.5px; }
.attention { color: var(--on-dim-soft); font-weight: 600; }

.controls { gap: 8px; }
.controls .toggled { color: var(--on-fact-soft); background: var(--fact-soft); }
.auto { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; }

.content {
  max-width: 1120px;
  width: 100%;
  margin: 0 auto;
  padding: 20px;
  flex: 1 1 auto;
}
.section-head { margin-bottom: 14px; }
.section-head h2 { margin: 0; font-size: 15px; color: var(--heading); }
.section-head p { margin: 2px 0 0; font-size: 12.5px; }

.list { display: flex; flex-direction: column; gap: 10px; }
.empty-state { padding: 28px; text-align: center; }
.foot { margin-top: 28px; font-size: 12px; text-align: center; }
</style>
