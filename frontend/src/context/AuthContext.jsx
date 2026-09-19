import { createContext, useState, useEffect } from 'react'
import { authApi } from '../api/auth'

export const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [token, setToken] = useState(() => localStorage.getItem('livepoll_jwt'))
  const [loading, setLoading] = useState(true)

  // Verify stored token on initial load
  useEffect(() => {
    async function verifyAuth() {
      const storedToken = localStorage.getItem('livepoll_jwt')
      if (!storedToken) {
        setUser(null)
        setLoading(false)
        return
      }

      try {
        const userData = await authApi.getMe()
        const storedName = localStorage.getItem('livepoll_user_name')
        if (storedName) {
          userData.name = storedName
        } else if (userData?.email) {
          userData.name = userData.email.split('@')[0]
        }
        setUser(userData)
        setToken(storedToken)
      } catch (err) {
        if (err.status === 401) {
          console.warn('Stored token is invalid or expired, clearing session.')
          localStorage.removeItem('livepoll_jwt')
          setUser(null)
          setToken(null)
        } else {
          // If backend was temporarily down (ERR_CONNECTION_REFUSED), keep token in case it restarts
          console.warn('Backend currently unreachable during auth check:', err.message)
          setUser(null)
        }
      } finally {
        setLoading(false)
      }
    }

    verifyAuth()

    // Listen for global 401 unauthorized notifications from apiClient
    const handleSessionExpired = (e) => {
      console.warn('Session expired event received:', e.detail?.message)
      localStorage.removeItem('livepoll_jwt')
      setUser(null)
      setToken(null)
    }

    window.addEventListener('livepoll:session_expired', handleSessionExpired)
    return () => {
      window.removeEventListener('livepoll:session_expired', handleSessionExpired)
    }
  }, [])

  const login = async (email, password) => {
    const data = await authApi.login(email, password)
    const storedName = localStorage.getItem('livepoll_user_name')
    if (storedName) {
      data.user.name = storedName
    } else if (data.user?.email) {
      data.user.name = data.user.email.split('@')[0]
    }
    localStorage.setItem('livepoll_jwt', data.token)
    setToken(data.token)
    setUser(data.user)
    return data
  }

  const signup = async (email, password, name = '') => {
    const data = await authApi.signup(email, password)
    if (name.trim()) {
      localStorage.setItem('livepoll_user_name', name.trim())
      data.user.name = name.trim()
    } else if (data.user?.email) {
      data.user.name = data.user.email.split('@')[0]
    }
    localStorage.setItem('livepoll_jwt', data.token)
    setToken(data.token)
    setUser(data.user)
    return data
  }

  const logout = () => {
    localStorage.removeItem('livepoll_jwt')
    localStorage.removeItem('livepoll_user_name')
    setToken(null)
    setUser(null)
  }

  const value = {
    user,
    token,
    loading,
    isAuthenticated: !!token && !!user,
    login,
    signup,
    logout,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
