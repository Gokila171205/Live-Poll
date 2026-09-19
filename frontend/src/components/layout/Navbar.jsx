import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { PollFlowLogo } from '../common/PollFlowLogo'
import { Button } from '../common/Button'
import { PlusIcon, UserIcon } from '../icons/Icons'

export function Navbar() {
  const { user, isAuthenticated, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  // Determine if this is a public voter page (keep voter screen distraction-free)
  const isVoterPage =
    /^\/(polls|poll|vote)\/[^/]+(?:\/vote)?$/.test(location.pathname) &&
    !location.pathname.includes('/results') &&
    !location.pathname.includes('/manage')

  if (isVoterPage) {
    return null
  }

  const handleLogout = () => {
    logout()
    navigate('/')
    setMobileMenuOpen(false)
  }

  const isActive = (path) => location.pathname === path

  const scrollToSection = (id) => {
    setMobileMenuOpen(false)
    if (location.pathname === '/') {
      const el = document.getElementById(id)
      if (el) {
        el.scrollIntoView({ behavior: 'smooth' })
        return
      }
    }
    navigate(`/#${id}`)
  }

  return (
    <header className="pollflow-navbar-wrapper">
      <div className="pollflow-navbar-container">
        {/* Left: PollFlow Brand Logo */}
        <Link to="/" className="navbar-brand-link" onClick={() => setMobileMenuOpen(false)}>
          <PollFlowLogo height={28} />
        </Link>

        {/* Center Navigation */}
        <nav className="navbar-center-links">
          <button
            type="button"
            onClick={() => scrollToSection('how-it-works')}
            className="navbar-nav-link"
          >
            How It Works
          </button>
          <button
            type="button"
            onClick={() => scrollToSection('features')}
            className="navbar-nav-link"
          >
            Features
          </button>
          <Link
            to={isAuthenticated ? '/dashboard' : '/login'}
            className={`navbar-nav-link ${isActive('/dashboard') ? 'active' : ''}`}
          >
            Dashboard
          </Link>
        </nav>

        {/* Right Actions */}
        <div className="navbar-right-actions">
          {isAuthenticated ? (
            <div className="navbar-auth-cluster">
              <div className="navbar-user-chip" title={user?.email}>
                <UserIcon size={15} />
                <span className="navbar-user-name">{user?.name || user?.email?.split('@')[0]}</span>
              </div>
              <Button variant="ghost" size="sm" onClick={handleLogout}>
                Log out
              </Button>
              <Link to="/create">
                <Button variant="primary" size="sm" icon={<PlusIcon size={15} />}>
                  Create Poll
                </Button>
              </Link>
            </div>
          ) : (
            <div className="navbar-auth-cluster">
              <Link to="/login" className="navbar-login-link">
                Login
              </Link>
              <Link to="/create">
                <Button variant="primary" size="sm" icon={<PlusIcon size={15} />}>
                  Create Poll
                </Button>
              </Link>
            </div>
          )}
        </div>

        {/* Mobile Hamburger Toggle */}
        <button
          type="button"
          className="navbar-hamburger-btn"
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          aria-label="Toggle navigation menu"
          aria-expanded={mobileMenuOpen}
        >
          <span className={`hamburger-line ${mobileMenuOpen ? 'open' : ''}`} />
          <span className={`hamburger-line ${mobileMenuOpen ? 'open' : ''}`} />
          <span className={`hamburger-line ${mobileMenuOpen ? 'open' : ''}`} />
        </button>
      </div>

      {/* Mobile Drawer Menu */}
      {mobileMenuOpen && (
        <div className="navbar-mobile-drawer">
          <button
            type="button"
            onClick={() => scrollToSection('how-it-works')}
            className="mobile-drawer-link"
          >
            How It Works
          </button>
          <button
            type="button"
            onClick={() => scrollToSection('features')}
            className="mobile-drawer-link"
          >
            Features
          </button>
          <Link
            to={isAuthenticated ? '/dashboard' : '/login'}
            className="mobile-drawer-link"
            onClick={() => setMobileMenuOpen(false)}
          >
            Dashboard
          </Link>

          <div className="mobile-drawer-divider" />

          {isAuthenticated ? (
            <div className="mobile-drawer-auth-block">
              <div className="mobile-drawer-user">
                <UserIcon size={16} />
                <span>{user?.name || user?.email}</span>
              </div>
              <Link
                to="/create"
                className="mobile-drawer-cta"
                onClick={() => setMobileMenuOpen(false)}
              >
                <Button variant="primary" size="md" className="w-full">
                  Create Poll
                </Button>
              </Link>
              <Button variant="ghost" size="sm" onClick={handleLogout} className="w-full">
                Log out
              </Button>
            </div>
          ) : (
            <div className="mobile-drawer-auth-block">
              <Link
                to="/login"
                className="mobile-drawer-link"
                onClick={() => setMobileMenuOpen(false)}
              >
                Login
              </Link>
              <Link
                to="/create"
                className="mobile-drawer-cta"
                onClick={() => setMobileMenuOpen(false)}
              >
                <Button variant="primary" size="md" className="w-full">
                  Create Poll
                </Button>
              </Link>
            </div>
          )}
        </div>
      )}
    </header>
  )
}
