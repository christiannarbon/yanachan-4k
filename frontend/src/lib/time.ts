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

/*
 * The dashboard's formatting.
 *
 * The activity chart is keyed by YYYY-MM-DD strings, which the backend cut in
 * the server's own zone. `new Date("2026-08-24")` would read those as UTC
 * midnight, so west of Greenwich every column would render as the day before
 * -- parseLocalDate keeps a calendar day a calendar day.
 */

/** A YYYY-MM-DD day, read as local midnight rather than as UTC. */
export function parseLocalDate(date: string | null | undefined): Date | null {
  if (!date) return null
  const [y, m, d] = date.split('-').map(Number)
  if (!y || !m || !d) return null
  const out = new Date(y, m - 1, d)
  return Number.isNaN(out.getTime()) ? null : out
}

/** A chart axis tick: "Mon", "月". */
export function weekday(date: string | null | undefined, t: TimeMessages): string {
  const d = parseLocalDate(date)
  return d ? d.toLocaleDateString(t.dateLocale, t.weekdayFormat) : ''
}

/** A calendar day without the time: "24 Aug", "8月24日". */
export function day(date: string | null | undefined, t: TimeMessages): string {
  const d = parseLocalDate(date)
  return d ? d.toLocaleDateString(t.dateLocale, t.dayFormat) : ''
}

/** The same, from a full timestamp -- what the week's range is cut from. */
export function dayOf(iso: string | null | undefined, t: TimeMessages): string {
  if (!iso) return ''
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '' : d.toLocaleDateString(t.dateLocale, t.dayFormat)
}

/** Grouped digits: "1,204", "1,204". Figures are read, so they get separators. */
export function num(n: number, t: TimeMessages): string {
  return n.toLocaleString(t.dateLocale)
}
