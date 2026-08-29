/*
 * English message catalog -- the source of truth for the app's copy.
 *
 * Every string the UI renders lives here so ja.ts can mirror it key for key:
 * `Messages` is derived from this object, which turns a missing or misshapen
 * Japanese key into a type error rather than a hole in the interface.
 *
 * Two shapes appear below. A plain string is rendered as-is; a function takes
 * the values it interpolates, so a translation can put them where its own
 * grammar wants them. Sentences that need inline markup are the exception:
 * they keep {name} placeholders and are rendered through Msg.vue, whose named
 * slots let a translation reorder the marked-up pieces too.
 */

const en = {
  /** Product name. A proper noun, so it is not translated. */
  appName: 'Yana-chan 4K',

  common: {
    loading: 'Loading.',
    somethingWrong: 'Something went wrong.',
    remove: 'Remove',
    copy: 'Copy',
  },

  time: {
    justNow: 'just now',
    minutes: (n: number) => `${n}m ago`,
    hours: (n: number) => `${n}h ago`,
    days: (n: number) => `${n}d ago`,
    /** BCP-47 tag for Intl. Undefined follows the browser's own preference. */
    dateLocale: undefined as string | undefined,
    /** Cutoff stamp: "Fri 28 Aug 09:12", matching the script's cutoff line. */
    dateFormat: {
      weekday: 'short',
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    } as Intl.DateTimeFormatOptions,
    /** Between actor handles in a list. */
    listSeparator: ', ',
  },

  /** Labels for the activity window, keyed by the backend's window.kind. */
  window: {
    fixed: (hours: number) => `last ${hours}h`,
    businessDay: 'since last business day (Fri)',
  },

  /** Titles for the built-in tabs. Team and org tabs are named by their ref. */
  sections: {
    mine: 'Your open PRs',
    review: 'Review requested from you',
    settings: 'Settings',
  },

  board: {
    /** Topbar line under the title. Rendered through Msg.vue. */
    summary: '{login} · window: {label} (since {stamp})',
    needingAttention: (n: number) => `${n} needing attention`,
    prCount: (n: number) => `${n} pull request${n === 1 ? '' : 's'}`,
    hiddenAsQuiet: (n: number) => `${n} hidden as quiet`,
    showingActiveOnly: 'Showing active only',
    showingEverything: 'Showing everything',
    autoRefresh: 'Auto refresh',
    refresh: 'Refresh',
    refreshing: 'Refreshing',
    githubReported: (warning: string) => `GitHub reported: ${warning}`,
    empty: 'Nothing here.',
    /** Follows `empty` on a team or org tab. Rendered through Msg.vue. */
    emptyScope: 'This tab covers {ref}.',
    footer: (generated: string, limit: number) =>
      `Generated ${generated} · ${limit} pull requests per query`,
  },

  pr: {
    statusReply: 'Reply',
    statusNew: 'New',
    draft: 'Draft',
    checksSuccess: 'Checks passing',
    checksFailure: 'Checks failing',
    checksPending: 'Checks running',
    approved: 'Approved',
    changesRequested: 'Changes requested',
    reviewRequired: 'Review required',
    youApproved: 'You approved',
    youRequestedChanges: 'You requested changes',
    youCommented: 'You commented',
    alsoRequestedFromYou: 'Also requested from you',
    awaitingYou: 'Awaiting you',
    awaitingTeam: 'Awaiting team review',
    answered: (actors: string) => `${actors} answered`,
    bots: (actors: string) => `${actors} (bot)`,
    commentsFrom: (n: number, actors: string) =>
      `${n} comment${n > 1 ? 's' : ''} from ${actors}`,
    botUpdates: (n: number) => `${n} bot update${n > 1 ? 's' : ''}`,
    quietSince: (last: string) => `No new comments · last activity ${last}`,
    noComments: 'No comments yet',
    noActivityInWindow: 'No new activity in the window',
    byline: (author: string, opened: string) => ` · by @${author} · opened ${opened}`,
    /** Separator between the parts of the activity line. */
    partSeparator: ' · ',
  },

  auth: {
    intro:
      'Choose how the dashboard should talk to GitHub. Nothing is read until you approve one of these.',

    cliTitle: 'Use the gh CLI on this machine',
    cliDetected: 'Detected',
    cliNotLoggedIn: 'Not logged in',
    cliNotFound: 'Not found',
    /** Rendered through Msg.vue: {path}, {host} and {login} carry markup. */
    cliSession: 'gh is installed at {path} and logged in to {host} as {login}.',
    /** The same, for a session GitHub reports without a login. */
    cliSessionNoLogin: 'gh is installed at {path} and logged in to {host}.',
    cliExplain:
      'If you approve, the dashboard reads the token backing that session with {command} and uses it for every GitHub API call. The token stays on this machine, in the state directory, and is sent only to GitHub.',
    cliApprove: 'Approve and use my gh CLI session',
    cliChecking: 'Checking the token',
    cliNoAccount: 'gh is installed but no account is logged in.',
    cliRunLogin: 'Run {command} in a terminal, then reload this page.',
    cliMissing: 'gh CLI was not found on PATH.',
    cliContainerHint:
      'This is expected when the dashboard runs inside a container. Use the OAuth option below, or pass a token through {variable}.',

    envTitle: 'Use the token supplied by the environment',
    envAvailable: 'Available',
    envExplain:
      'A token was passed in through {variable}. This is how {command} forwards your local gh session into the container. Approve it to use that token.',
    envApprove: 'Approve and use the supplied token',

    oauthTitle: 'Sign in with GitHub',
    oauthEnabled: 'OAuth device flow',
    oauthNotConfigured: 'Not configured',
    oauthSetup:
      'Set {variable} to the client ID of an OAuth app with the device flow enabled, then restart the backend.',
    oauthScopes:
      'Requests the scopes {scopes}. GitHub shows you a code to enter in your browser; no client secret or callback URL is involved.',
    oauthStart: 'Start device sign-in',
    oauthContacting: 'Contacting GitHub',
    oauthOpen: 'Open {link} and enter this code:',
    oauthWaiting: 'Waiting for you to approve the code on github.com',
    oauthApproved: 'Approved',
    oauthExpired: 'The device code expired before it was approved. Start again.',
    oauthDenied: 'Authorization was denied on github.com.',
    oauthCopied: 'Code copied to the clipboard',
    oauthCopyManually: 'Copy the code manually',
  },

  settings: {
    saved: 'Saved. The board reloads with the new tabs.',
    teamFormat: 'A team must be written as org/team-slug.',
    orgFormat: 'An organization is a bare login, without a slash.',

    teamsTitle: 'Teams to follow',
    /** Rendered through Msg.vue: {format} carries markup. */
    teamsExplain:
      'Each team becomes its own tab, listing open pull requests where a review is requested from that team. Write them as {format}. Nothing is followed until you add it here.',
    teamsEmpty: 'No teams followed.',
    teamAdd: 'Add team',
    teamsYours: 'Your teams:',

    orgsTitle: 'Organizations to follow',
    orgsExplain:
      'Each organization becomes a tab listing open pull requests in that org that involve you and are not already shown on another tab.',
    orgsEmpty: 'No organizations followed.',
    orgAdd: 'Add organization',
    orgsYours: 'Your organizations:',
    membershipsLoading: 'Loading your memberships.',
    membershipsFailed: (warning: string) => `Could not list your memberships: ${warning}`,

    viewTitle: 'View',
    limitLabel: 'Pull requests per query',
    /** Rendered through Msg.vue: {flag} carries markup. */
    limitHint: "1 to 100, matching the script's {flag}.",
    windowLabel: 'Activity window',
    windowHint: 'Hours. Leave 0 to use the business-day rule: Mon looks back to Friday.',
    onlyActive: 'Hide pull requests with no new activity in the window',
    showUrls: 'Show the full URL under each pull request',
    save: 'Save settings',
    saving: 'Saving',

    sessionTitle: 'Session',
    /** Rendered through Msg.vue: {login} carries markup. */
    sessionLine: 'Signed in as {login} using the {mode}.',
    modeGhCli: 'local gh CLI session',
    modeOauth: 'OAuth device flow',
    modeEnvToken: 'token from the environment',
    signOut: 'Sign out and forget the token',
  },

  theme: {
    label: 'Theme',
    current: (name: string) => `Theme: ${name}`,
    groupOde: 'Ode to Yanami',
    groupHouse: 'House',
    groupPainting: 'Paintings',
  },

  locale: {
    label: 'Language',
    /** Each locale names itself, so both read natively in either language. */
    en: 'English',
    ja: '日本語',
  },
}

/** The shape every catalog must fill. */
export type Messages = typeof en

export default en
