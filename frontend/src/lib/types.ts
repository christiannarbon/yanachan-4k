export interface GhCliStatus {
  available: boolean
  authenticated: boolean
  login: string
  host: string
  path: string
  detail?: string
}

export interface SessionView {
  mode: 'gh-cli' | 'oauth' | 'env-token' | string
  login: string
}

export interface AuthStatus {
  authenticated: boolean
  session: SessionView | null
  ghCli: GhCliStatus
  ghCliAllowed: boolean
  oauthEnabled: boolean
  envTokenAvailable: boolean
  oauthScopes: string
  pendingDevice?: { expiresAt: string; interval: number }
}

export interface DeviceStart {
  userCode: string
  verificationUri: string
  expiresIn: number
  interval: number
}

export interface DevicePoll {
  state: 'pending' | 'slow_down' | 'complete' | 'expired' | 'denied'
  session?: SessionView
}

export interface Settings {
  teams: string[]
  orgs: string[]
  limit: number
  windowHours: number
  onlyActive: boolean
  showUrls: boolean
}

export type EntryStatus = 'reply' | 'new' | 'quiet' | ''

export interface Entry {
  number: number
  title: string
  url: string
  repo: string
  isDraft: boolean
  author: string
  authorIsBot: boolean
  createdAt: string
  updatedAt: string
  checks: '' | 'success' | 'failure' | 'pending'
  reviewDecision: '' | 'approved' | 'changes_requested' | 'review_required'
  active: boolean
  hot: boolean
  status: EntryStatus
  humanActors: string[]
  botActors: string[]
  humanCount: number
  botCount: number
  lastActivityAt: string | null
  touched: boolean
  yourState: '' | 'approved' | 'changes_requested' | 'commented'
  yourLastAt: string | null
  awaiting: '' | 'you' | 'team'
  alsoRequestedFromYou: boolean
}

export interface Section {
  id: string
  title: string
  kind: 'mine' | 'review' | 'team' | 'org'
  ref: string
  entries: Entry[]
  total: number
  active: number
  hot: number
  error?: string
}

export interface WindowView {
  /** Machine-readable, so the frontend can word the window in any language. */
  kind: 'fixed' | 'business-day' | string
  /** The backend's own English wording, used as the fallback. */
  label: string
  hours: number
  cutoff: string
  now: string
}

export interface Board {
  login: string
  authMode: string
  window: WindowView
  sections: Section[]
  generatedAt: string
  onlyActive: boolean
  limit: number
  warning?: string
}

export interface Suggestions {
  orgs: string[]
  teams: string[]
  warning?: string
}
