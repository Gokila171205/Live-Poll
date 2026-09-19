/**
 * URL resolution helpers for LivePoll frontend.
 * Ensures share links and QR codes always use the correct frontend domain
 * (Vercel deployment URL in production, or window.location.origin, never the backend URL).
 */

export function getFrontendBaseUrl() {
  const envUrl = import.meta.env.VITE_APP_URL || import.meta.env.VITE_PUBLIC_URL
  if (envUrl) {
    return envUrl.replace(/\/+$/, '')
  }
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin
  }
  return 'https://live-poll-ashy.vercel.app'
}

/**
 * Returns the public voter URL for a poll.
 * Generates /poll/:id (also compatible with /polls/:id in router).
 */
export function getPollShareUrl(pollId) {
  return `${getFrontendBaseUrl()}/poll/${pollId}`
}

/**
 * Returns the live results presentation screen URL for a poll.
 */
export function getPollResultsUrl(pollId) {
  return `${getFrontendBaseUrl()}/poll/${pollId}/results`
}
