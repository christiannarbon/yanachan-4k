/*
 * Time and actor formatting.
 *
 * The wording lives in the message catalogs, so these take the catalog's
 * `time` section and decide only which of its entries applies. Components
 * rarely call them directly -- useI18n() hands back versions already bound to
 * the active locale.
 */

import type { Messages } from '../i18n/en'

type TimeMessages = Messages['time']

/** Relative time in the same shape the shell script used: "3m ago", "2h ago". */
export function ago(iso: string | null | undefined, now: number, t: TimeMessages): string {
  if (!iso) return ''
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const seconds = Math.max(0, Math.floor((now - then) / 1000))
  if (seconds < 60) return t.justNow
  if (seconds < 3600) return t.minutes(Math.floor(seconds / 60))
  if (seconds < 86400) return t.hours(Math.floor(seconds / 3600))
  return t.days(Math.floor(seconds / 86400))
}

/** The cutoff line's stamp: "Fri 28 Aug 09:12", or "8月28日(金) 09:12". */
export function stamp(iso: string | null | undefined, t: TimeMessages): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(t.dateLocale, t.dateFormat)
}

/** Joins logins the way the jq "actors" helper did: at most three, then "+N". */
export function actorList(logins: string[], separator: string, max = 3): string {
  if (logins.length === 0) return ''
  if (logins.length <= max) return logins.map((l) => `@${l}`).join(separator)
  return [...logins.slice(0, max).map((l) => `@${l}`), `+${logins.length - max}`].join(separator)
}
