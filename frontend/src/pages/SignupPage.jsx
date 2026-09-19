import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { Input } from '../components/common/Input'
import { Button } from '../components/common/Button'
import { Alert } from '../components/common/Alert'
import { PollFlowLogo } from '../components/common/PollFlowLogo'
import { MailIcon, LockIcon, UserIcon, EyeIcon, EyeOffIcon, CheckCircleIcon } from '../components/icons/Icons'

export function SignupPage() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(false)

  const { signup } = useAuth()
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)

    if (!name.trim()) {
      setError('Please enter your name.')
      return
    }

    if (!email.trim() || !password) {
      setError('Please provide an email address and password.')
      return
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters long.')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match.')
      return
    }

    setLoading(true)
    try {
      await signup(email.trim(), password, name.trim())
      navigate('/dashboard')
    } catch (err) {
      setError(err.message || 'Failed to create account. Please try again.')
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
          <h1 className="auth-page-heading">Create your PollFlow account</h1>
          <p className="auth-page-subheading">
            Start creating real-time polls and engaging your audience for free.
          </p>
        </div>

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
            label="Full name"
            type="text"
            placeholder="Sarah Jenkins"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            icon={<UserIcon size={16} />}
          />

          <Input
            label="Work or personal email"
            type="email"
            placeholder="you@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            icon={<MailIcon size={16} />}
          />

          <Input
            label="Password"
            type={showPassword ? 'text' : 'password'}
            placeholder="At least 8 characters"
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

          <Input
            label="Confirm password"
            type={showConfirmPassword ? 'text' : 'password'}
            placeholder="Re-enter your password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            icon={<LockIcon size={16} />}
            suffix={
              <button
                type="button"
                className="password-reveal-btn"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                tabIndex={-1}
                title={showConfirmPassword ? 'Hide password' : 'Show password'}
              >
                {showConfirmPassword ? <EyeOffIcon size={16} /> : <EyeIcon size={16} />}
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
            Create Account
          </Button>
        </form>

        <div className="auth-card-bottom-row text-center">
          <span>Already have an account?</span>{' '}
          <Link to="/login" className="auth-switch-link">
            Log in
          </Link>
        </div>

        <div className="auth-trust-strip">
          <div className="trust-strip-item">
            <CheckCircleIcon size={14} className="trust-check" />
            <span>No credit card required</span>
          </div>
          <div className="trust-strip-item">
            <CheckCircleIcon size={14} className="trust-check" />
            <span>Instant live setup</span>
          </div>
        </div>
      </div>
    </div>
  )
}
