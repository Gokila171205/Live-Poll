import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { LoadingSpinner } from '../common/LoadingSpinner'

export function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="page-loading-wrapper">
        <LoadingSpinner text="Checking authentication session..." size="lg" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <Navigate
        to="/login"
        state={{
          from: location,
          message: 'Please sign in to access your creator dashboard.',
        }}
        replace
      />
    )
  }

  return children
}
