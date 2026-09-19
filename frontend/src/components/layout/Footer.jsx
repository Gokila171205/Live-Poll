import { Link, useLocation } from 'react-router-dom'
import { PollFlowLogo } from '../common/PollFlowLogo'

export function Footer() {
  const location = useLocation()

  // Hide footer on workspace/dashboard, live results, and public voter page
  const isAppWorkspace =
    location.pathname.startsWith('/dashboard') ||
    location.pathname.includes('/results') ||
    (/^\/(polls|poll|vote)\/[^/]+(?:\/vote)?$/.test(location.pathname) && !location.pathname.includes('/manage'))

  if (isAppWorkspace) {
    return null
  }

  return (
    <footer className="pollflow-footer">
      <div className="footer-inner">
        <div className="footer-content-grid">
          <div className="footer-brand-col">
            <Link to="/" className="footer-brand-link">
              <PollFlowLogo height={26} />
            </Link>
            <p className="footer-tagline">
              Create polls. Share instantly. See results live.
            </p>
            <p className="footer-mission">
              Modern real-time polling platform designed for presenters, educators, and teams.
            </p>
          </div>

          <div className="footer-links-grid">
            <div className="footer-nav-col">
              <span className="footer-heading">Product</span>
              <Link to="/create" className="footer-nav-item">Create Poll</Link>
              <a href="/#how-it-works" className="footer-nav-item">How It Works</a>
              <a href="/#features" className="footer-nav-item">Features</a>
              <Link to="/dashboard" className="footer-nav-item">Dashboard</Link>
            </div>

            <div className="footer-nav-col">
              <span className="footer-heading">Platform</span>
              <span className="footer-nav-item muted">Go WebSocket Engine</span>
              <span className="footer-nav-item muted">Redis Pub/Sub Sync</span>
              <span className="footer-nav-item muted">Sub-second Latency</span>
            </div>

            <div className="footer-nav-col">
              <span className="footer-heading">Account</span>
              <Link to="/login" className="footer-nav-item">Sign In</Link>
              <Link to="/signup" className="footer-nav-item">Get Started</Link>
            </div>
          </div>
        </div>

        <div className="footer-bottom-bar">
          <p className="footer-copyright">
            © {new Date().getFullYear()} PollFlow. All rights reserved.
          </p>
          <div className="footer-legal-links">
            <span>Privacy</span>
            <span className="dot-sep">·</span>
            <span>Terms</span>
            <span className="dot-sep">·</span>
            <span className="system-status-indicator">
              <span className="live-status-dot" /> System Operational
            </span>
          </div>
        </div>
      </div>
    </footer>
  )
}
