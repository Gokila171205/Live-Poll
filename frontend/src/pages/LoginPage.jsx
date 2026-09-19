import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { Input } from '../components/common/Input'
import { Button } from '../components/common/Button'
import { Alert } from '../components/common/Alert'
import { PollFlowLogo } from '../components/common/PollFlowLogo'
import { MailIcon, LockIcon, EyeIcon, EyeOffIcon, CheckCircleIcon } from '../components/icons/Icons'

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(false)

  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const infoMessage = location.state?.message || null

  const from = location.state?.from?.pathname || '/dashboard'

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)

    if (!email.trim() || !password) {
      setError('Please provide both your email address and password.')
      return
    }

    setLoading(true)
    try {
      await login(email.trim(), password)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err.message || 'Invalid email or password.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page-container">
      <div className="auth-card-shell">
        <div className="auth-card-header text-center">
          <Link to="/" className="auth-logo-link">
            <PollFlowLogo height={32} />
          </Link>
          <h1 className="auth-page-heading">Welcome back</h1>
          <p className="auth-page-subheading">
            Log in to manage your real-time polls and present live.
          </p>
        </div>

        {infoMessage && (
          <Alert type="info" message={infoMessage} className="mb-20" />
        )}

        {error && (
          <Alert
            type="error"
            message={error}
            onClose={() => setError(null)}
            className="mb-20"
          />
        )}

        <form onSubmit={handleSubmit} className="auth-form-body">
          <Input
            label="Email address"
            type="email"
            placeholder="name@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            icon={<MailIcon size={16} />}
          />

          <Input
            label="Password"
            type={showPassword ? 'text' : 'password'}
            placeholder="Enter your account password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            icon={<LockIcon size={16} />}
            suffix={
              <button
                type="button"
                className="password-reveal-btn"
                onClick={() => setShowPassword(!showPassword)}
                tabIndex={-1}
                title={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOffIcon size={16} /> : <EyeIcon size={16} />}
              </button>
            }
          />

          <Button
            type="submit"
            variant="primary"
            size="lg"
            isLoading={loading}
            className="w-full auth-submit-btn mt-12"
          >
            Log In
          </Button>
        </form>

        <div className="auth-card-bottom-row text-center">
          <span>Don't have an account yet?</span>{' '}
          <Link to="/signup" className="auth-switch-link">
            Create an account
          </Link>
        </div>

        {/* Feature highlight strip to prevent visually empty feel */}
        <div className="auth-trust-strip">
          <div className="trust-strip-item">
            <CheckCircleIcon size={14} className="trust-check" />
            <span>Instant live results</span>
          </div>
          <div className="trust-strip-item">
            <CheckCircleIcon size={14} className="trust-check" />
            <span>Unlimited audience votes</span>
          </div>
        </div>
      </div>
    </div>
  )
}
