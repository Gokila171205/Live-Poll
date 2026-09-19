function getApiBase() {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL.replace(/\/+$/, '')
  }
  // If running locally in development, default to local backend
  if (typeof window !== 'undefined' && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')) {
    return 'http://localhost:8080/api'
  }
  // Production Render backend
  return 'https://live-poll-v9go.onrender.com/api'
}

const API_BASE = getApiBase()

/**
 * Universal fetch wrapper for LivePoll REST API.
 * Automatically injects authorization header if token is in localStorage.
 */
export async function apiClient(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`
  
  const token = localStorage.getItem('livepoll_jwt')
  
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  }

  const config = {
    ...options,
    headers,
  }

  let response
  try {
    response = await fetch(url, config)
  } catch (netErr) {
    if (typeof window !== 'undefined' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1' && url.includes('localhost:8080')) {
      throw new Error(`Unable to reach backend: The app is trying to connect to "${url}", which is unreachable from an external device. Please set VITE_API_BASE_URL to your deployed Render backend URL in your Vercel project settings.`)
    }
    throw new Error(`Network connection error: ${netErr.message}`)
  }

  // Parse JSON response safely
  let data = null
  const contentType = response.headers.get('content-type')
  if (contentType && contentType.includes('application/json')) {
    try {
      data = await response.json()
    } catch {
      data = null
    }
  }

  if (!response.ok) {
    if (response.status === 401) {
      // Clear stale/invalid credentials
      localStorage.removeItem('livepoll_jwt')
      if (typeof window !== 'undefined') {
        window.dispatchEvent(
          new CustomEvent('livepoll:session_expired', {
            detail: { message: data?.error?.message || 'Session expired or unauthorized.' }
          })
        )
      }
    }

    const errorMsg = data?.error?.message || data?.message || `Request failed with status ${response.status}`
    const error = new Error(errorMsg)
    error.status = response.status
    error.code = data?.error?.code || 'API_ERROR'
    error.data = data
    throw error
  }

  return data
}
