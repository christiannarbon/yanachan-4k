/**
 * Locale state: which catalog supplies the app's copy.
 *
 * There are two, English and Japanese. The first visit picks one from the
 * browser's language preferences; after that the choice persists per browser,
 * the same way a theme does -- see composables/useTheme.ts.
 *
 * Strings are looked up as plain property access on a catalog object rather
 * than through a key string, so a typo is a type error and vue-tsc catches a
 * key that exists in only one language. Sentences with inline markup keep
 * {name} placeholders and render through Msg.vue.
 */

import { computed, ref, watch } from 'vue'

import { actorList, ago, day, dayOf, num, stamp, weekday } from '../lib/time'
import type { Board, Section } from '../lib/types'
import en, { type Messages } from './en'
import ja from './ja'

export type Locale = 'en' | 'ja'

const CATALOGS: Record<Locale, Messages> = { en, ja }

/** Picker order. Each catalog names its own language. */
export const LOCALES: Locale[] = ['en', 'ja']

const LOCALE_KEY = 'yana.locale'
const DEFAULT_LOCALE: Locale = 'en'

function isLocale(v: unknown): v is Locale {
  return typeof v === 'string' && (LOCALES as string[]).includes(v)
}

/** First supported language among the browser's preferences, else English. */
function detectLocale(): Locale {
  if (typeof navigator === 'undefined') return DEFAULT_LOCALE
  const prefs = navigator.languages?.length ? navigator.languages : [navigator.language]
  for (const pref of prefs) {
    // Preferences arrive as full tags: ja, ja-JP, en-GB.
    const base = pref?.toLowerCase().split('-')[0]
    if (isLocale(base)) return base
  }
  return DEFAULT_LOCALE
}

function readStoredLocale(): Locale {
  try {
    const v = localStorage.getItem(LOCALE_KEY)
    if (isLocale(v)) return v
  } catch {
    // Private windows and blocked site data both throw; fall through.
  }
  return detectLocale()
}

const locale = ref<Locale>(readStoredLocale())

/**
 * Keeps <html lang> in step with the catalog, which is what tells the browser
 * to pick a Japanese face for the text and hyphenate and quote it correctly.
 */
function apply(v: Locale) {
  if (typeof document === 'undefined') return
  document.documentElement.setAttribute('lang', v)
}

apply(locale.value)

watch(locale, (v) => {
  apply(v)
  try {
    localStorage.setItem(LOCALE_KEY, v)
  } catch {
    // Persisting is a convenience; the choice still applies for this session.
  }
})

export function useI18n() {
  /** The active catalog. Templates read it as `t.board.refresh`. */
  const t = computed(() => CATALOGS[locale.value])

  /** Language names for the picker, each written in its own language. */
  const localeNames = computed(() =>
    LOCALES.map((id) => ({ id, name: CATALOGS[id].locale[id] })),
  )

  function setLocale(v: Locale) {
    locale.value = v
  }

  // The helpers below read t.value when called rather than closing over a
  // catalog, so a component that only formats a date still re-renders when the
  // language changes.

  return {
    locale,
    locales: LOCALES,
    localeNames,
    t,
    setLocale,

    /** "3m ago" / "3 分前". */
    ago: (iso: string | null | undefined, now: number) => ago(iso, now, t.value.time),
    /** "Fri 28 Aug 09:12" / "8月28日(金) 09:12". */
    stamp: (iso: string | null | undefined) => stamp(iso, t.value.time),
    /** "@ada, @grace, +2", joined the way the language punctuates a list. */
    actors: (logins: string[]) => actorList(logins, t.value.time.listSeparator),
    /** A chart axis tick from a YYYY-MM-DD day: "Mon" / "月". */
    weekday: (date: string | null | undefined) => weekday(date, t.value.time),
    /** A YYYY-MM-DD day written out: "24 Aug" / "8月24日". */
    day: (date: string | null | undefined) => day(date, t.value.time),
    /** The same, from a full timestamp. */
    dayOf: (iso: string | null | undefined) => dayOf(iso, t.value.time),
    /** A figure with digit grouping: "1,204". */
    num: (n: number) => num(n, t.value.time),

    /**
     * A tab's heading. The built-in tabs are translated; a team or org tab is
     * named by its ref -- an identifier, the same in every language -- and the
     * backend's own English title is the fallback for a kind added later.
     */
    sectionTitle: (section: Pick<Section, 'kind' | 'ref' | 'title'>) => {
      switch (section.kind) {
        case 'mine':
          return t.value.sections.mine
        case 'review':
          return t.value.sections.review
        case 'team':
        case 'org':
          return section.ref
        default:
          return section.title
      }
    },

    /**
     * The activity window, from the backend's machine-readable kind. Falls
     * back to the English label the backend also sends.
     */
    windowLabel: (w: Board['window']) => {
      switch (w.kind) {
        case 'fixed':
          return t.value.window.fixed(w.hours)
        case 'business-day':
          return t.value.window.businessDay
        default:
          return w.label
      }
    },
  }
}
