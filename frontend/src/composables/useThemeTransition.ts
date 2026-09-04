/**
 * Runs a theme change as a view transition.
 *
 * The browser snapshots the page, we swap the palette, and it holds the old
 * frame in place while the new one wipes in from the left -- the animation
 * itself lives in styles/theme-transition.css. Nothing in a component has to
 * know: the whole page is the transition.
 *
 * Two ways out, both silent. A browser without view transitions gets the swap
 * on its own, exactly as before this existed; so does anyone who has asked for
 * reduced motion, and that is checked at the moment of the change rather than
 * once at load, so flipping the system setting takes effect immediately.
 */

/** Present on <html> only while a theme sweep runs, so the stylesheet's rules
 *  cannot catch a view transition started by anything else later. */
const SWEEP_ATTR = 'data-theme-sweep'

function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * Applies `swap` -- whatever changes the palette -- with the sweep around it.
 *
 * `swap` may be async: a view transition waits on the promise it returns
 * before taking the second snapshot, which is how a Vue watcher gets a chance
 * to flip the attribute inside the transition rather than after it.
 */
export function sweepThemeChange(swap: () => void | Promise<void>): void {
  if (typeof document === 'undefined') return void swap()

  if (typeof document.startViewTransition !== 'function' || prefersReducedMotion()) {
    void swap()
    return
  }

  const root = document.documentElement
  root.setAttribute(SWEEP_ATTR, '')
  const transition = document.startViewTransition(swap)
  // `finished` rejects if the swap throws, and a skipped transition still
  // settles, so this is the one place the attribute comes off.
  void transition.finished
    .catch(() => {})
    .finally(() => root.removeAttribute(SWEEP_ATTR))
}
