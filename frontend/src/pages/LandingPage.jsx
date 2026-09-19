import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { Button } from '../components/common/Button'
import {
  PlusIcon,
  ArrowRightIcon,
  CheckCircleIcon,
  ShareIcon,
  ChartIcon,
  ChevronRightIcon,
} from '../components/icons/Icons'

export function LandingPage() {
  const { isAuthenticated } = useAuth()

  // Interactive Demo Poll State for Hero Product Preview
  // Required: "Which technology do you enjoy working with most?" with React, Python, Java, Go
  // Initial total votes: 324
  const [demoOptions, setDemoOptions] = useState([
    { id: 'react', text: 'React', votes: 156 },
    { id: 'python', text: 'Python', votes: 104 },
    { id: 'java', text: 'Java', votes: 44 },
    { id: 'go', text: 'Go', votes: 20 },
  ])
  const [selectedDemoId, setSelectedDemoId] = useState(null)

  const totalDemoVotes = demoOptions.reduce((acc, opt) => acc + opt.votes, 0)

  const handleDemoVote = (optionId) => {
    if (selectedDemoId === optionId) return
    setDemoOptions((prev) =>
      prev.map((opt) => {
        if (opt.id === optionId) {
          return { ...opt, votes: opt.votes + 1 }
        }
        if (selectedDemoId && opt.id === selectedDemoId) {
          return { ...opt, votes: Math.max(0, opt.votes - 1) }
        }
        return opt
      })
    )
    setSelectedDemoId(optionId)
  }

  const scrollToHowItWorks = (e) => {
    e.preventDefault()
    const element = document.getElementById('how-it-works')
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' })
    }
  }

  return (
    <div className="landing-page-root">
      {/* ====================================================================
          1. HERO SECTION
          Headline: "Create polls. Share instantly. See results live."
          Supporting: "Create engaging polls, share them with your audience, and watch responses update in real time."
          Buttons: "Create a Poll", "See How It Works"
          Hero illustration: /images/hero-poll.png
          ==================================================================== */}
      <section className="hero-section">
        <div className="hero-container">
          <div className="hero-text-col">
            <div className="hero-eyebrow-badge">
              <span className="live-pulsing-dot" />
              <span>Real-Time Polling Platform</span>
            </div>

            <h1 className="hero-heading">
              Create polls.
              <br />
              Share instantly.
              <br />
              <span className="hero-heading-highlight">See results live.</span>
            </h1>

            <p className="hero-description">
              Create engaging polls, share them with your audience, and watch responses update in real time.
            </p>

            <div className="hero-cta-cluster">
              <Link to={isAuthenticated ? '/create' : '/signup'}>
                <Button variant="primary" size="lg" icon={<PlusIcon size={18} />}>
                  Create a Poll
                </Button>
              </Link>
              <a href="#how-it-works" onClick={scrollToHowItWorks}>
                <Button variant="secondary" size="lg" icon={<ArrowRightIcon size={16} />}>
                  See How It Works
                </Button>
              </a>
            </div>

            <div className="hero-trust-row">
              <div className="trust-badge-item">
                <CheckCircleIcon size={16} className="trust-icon" />
                <span>Zero app installs</span>
              </div>
              <div className="trust-badge-item">
                <CheckCircleIcon size={16} className="trust-icon" />
                <span>Free instant setup</span>
              </div>
              <div className="trust-badge-item">
                <CheckCircleIcon size={16} className="trust-icon" />
                <span>Sub-second live sync</span>
              </div>
            </div>
          </div>

          <div className="hero-visual-col">
            <div className="hero-image-wrapper">
              <img
                src="/images/hero-poll.png"
                alt="PollFlow Real-time polling illustration showing creation, sharing and live tally"
                className="hero-responsive-image"
                loading="eager"
              />
            </div>
          </div>
        </div>
      </section>

      {/* ====================================================================
          2. HERO PRODUCT PREVIEW (Interactive Demo)
          "Which technology do you enjoy working with most?"
          React, Python, Java, Go
          ● LIVE, 324 votes
          ==================================================================== */}
      <section className="hero-preview-section">
        <div className="hero-preview-inner">
          <div className="preview-intro-header text-center">
            <span className="subtle-section-pill">LIVE PRODUCT PREVIEW</span>
            <h2 className="preview-section-title">Experience the real-time speed</h2>
            <p className="preview-section-desc">
              Try voting below to see how responses update instantly on screen.
            </p>
          </div>

          <div className="interactive-poll-preview-card">
            {/* Card Header with LIVE indicator and vote count */}
            <div className="preview-card-top-bar">
              <div className="preview-badge-status">
                <span className="live-status-pill">
                  <span className="live-dot-red" />
                  LIVE
                </span>
                <span className="preview-poll-brand">PollFlow Realtime</span>
              </div>
              <div className="preview-vote-badge">
                <strong>{totalDemoVotes}</strong> votes
              </div>
            </div>

            {/* Poll Question */}
            <h3 className="preview-question-title">
              Which technology do you enjoy working with most?
            </h3>
            <p className="preview-question-hint">
              Tap an option to test real-time percentage recalculation:
            </p>

            {/* Options List with Dynamic Bars */}
            <div className="preview-options-list">
              {demoOptions.map((option) => {
                const percentage =
                  totalDemoVotes > 0 ? Math.round((option.votes / totalDemoVotes) * 100) : 0
                const isSelected = selectedDemoId === option.id

                return (
                  <div
                    key={option.id}
                    className={`preview-option-item ${isSelected ? 'selected' : ''}`}
                    onClick={() => handleDemoVote(option.id)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') handleDemoVote(option.id)
                    }}
                  >
                    <div className="preview-option-content">
                      <div className="preview-option-left">
                        <span className={`custom-radio-circle ${isSelected ? 'checked' : ''}`}>
                          {isSelected && <span className="custom-radio-inner" />}
                        </span>
                        <span className="preview-option-label">{option.text}</span>
                        {isSelected && <span className="your-vote-tag">Your vote</span>}
                      </div>
                      <div className="preview-option-stats">
                        <span className="preview-percentage-text">{percentage}%</span>
                        <span className="preview-count-text">({option.votes} votes)</span>
                      </div>
                    </div>

                    {/* Animated Progress Bar */}
                    <div className="preview-bar-track">
                      <div
                        className="preview-bar-fill"
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                )
              })}
            </div>

            <div className="preview-card-footer">
              <span className="preview-powered-text">Powered by PollFlow Engine</span>
              <span className="preview-simulated-note">
                Redis Pub/Sub + WebSocket sub-second synchronization
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* ====================================================================
          3. HOW POLLFLOW WORKS
          01 — Create: Create your question and answer options.
          02 — Share: Share your unique poll link.
          03 — Vote: Let your audience vote from any device.
          04 — See Live Results: Watch responses update instantly.
          Use: /images/how-it-works.png
          Create a clean visual progression: Create → Share → Vote → Live Results
          ==================================================================== */}
      <section id="how-it-works" className="how-it-works-section">
        <div className="section-container">
          <div className="section-header-centered">
            <span className="subtle-section-pill">HOW IT WORKS</span>
            <h2 className="section-main-title">A simple four-step polling workflow</h2>
            <p className="section-main-subtitle">
              Designed for ease of use, speed, and high audience engagement.
            </p>
          </div>

          <div className="workflow-diagram-row">
            <div className="workflow-image-container">
              <img
                src="/images/how-it-works.png"
                alt="PollFlow 4-step workflow: Create, Share, Vote, See Live Results"
                className="workflow-responsive-image"
                loading="lazy"
              />
            </div>

            {/* Visual Step Cards with connecting hierarchy */}
            <div className="workflow-steps-vertical-grid">
              <div className="workflow-step-box">
                <div className="step-badge-number">01</div>
                <div className="step-info-block">
                  <h3 className="step-name">Create</h3>
                  <p className="step-explanation">
                    Create your question and answer options in seconds.
                  </p>
                </div>
              </div>

              <div className="step-connector-line">
                <ChevronRightIcon size={16} />
              </div>

              <div className="workflow-step-box">
                <div className="step-badge-number">02</div>
                <div className="step-info-block">
                  <h3 className="step-name">Share</h3>
                  <p className="step-explanation">
                    Share your unique poll link or present the instant QR code.
                  </p>
                </div>
              </div>

              <div className="step-connector-line">
                <ChevronRightIcon size={16} />
              </div>

              <div className="workflow-step-box">
                <div className="step-badge-number">03</div>
                <div className="step-info-block">
                  <h3 className="step-name">Vote</h3>
                  <p className="step-explanation">
                    Let your audience vote from any device with no logins or downloads.
                  </p>
                </div>
              </div>

              <div className="step-connector-line">
                <ChevronRightIcon size={16} />
              </div>

              <div className="workflow-step-box highlight-step">
                <div className="step-badge-number highlight">04</div>
                <div className="step-info-block">
                  <h3 className="step-name">See Live Results</h3>
                  <p className="step-explanation">
                    Watch responses update instantly with smooth animated bars.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ====================================================================
          4. FEATURES SECTION (Three Major Feature Showcases)
          1. CREATE POLLS: "Create polls in seconds."
             "Build clear, engaging questions with flexible answer options."
             Image: /images/create-poll.png
          2. SHARE ANYWHERE: "One link. Everyone can vote."
             "Share your poll with your audience using a simple public link."
             Image: /images/share-poll.png
          3. REAL-TIME RESULTS: "Watch results happen live."
             "See votes update instantly without refreshing the page."
             Image: /images/live-results.png
             Mention: Redis + WebSocket powered real-time updates
          ==================================================================== */}
      <section id="features" className="features-section">
        <div className="section-container">
          <div className="section-header-centered">
            <span className="subtle-section-pill">POWERFUL CAPABILITIES</span>
            <h2 className="section-main-title">Everything you need to run live polls</h2>
            <p className="section-main-subtitle">
              Professional features engineered for presentations, classrooms, webinars, and live streams.
            </p>
          </div>

          {/* Feature 1: CREATE POLLS */}
          <div className="feature-block-row">
            <div className="feature-text-side">
              <span className="feature-kicker-tag">CREATE POLLS</span>
              <h3 className="feature-headline">Create polls in seconds.</h3>
              <p className="feature-body-text">
                Build clear, engaging questions with flexible answer options. Customize voting rules, set single-response limits, and organize your polls in an intuitive workspace.
              </p>
              <ul className="feature-check-list">
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Dynamic option rows with instant validation</span>
                </li>
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Real-time synchronous preview as you type</span>
                </li>
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Configurable voting permissions and status toggles</span>
                </li>
              </ul>
              <div className="feature-cta-wrap">
                <Link to={isAuthenticated ? '/create' : '/signup'}>
                  <Button variant="primary" size="md" icon={<PlusIcon size={16} />}>
                    Build a Poll Now
                  </Button>
                </Link>
              </div>
            </div>
            <div className="feature-visual-side">
              <div className="feature-image-card">
                <img
                  src="/images/create-poll.png"
                  alt="Create Poll interface illustration"
                  className="feature-responsive-image"
                  loading="lazy"
                />
              </div>
            </div>
          </div>

          {/* Feature 2: SHARE ANYWHERE (Reversed) */}
          <div className="feature-block-row reverse">
            <div className="feature-visual-side">
              <div className="feature-image-card">
                <img
                  src="/images/share-poll.png"
                  alt="Share Poll link and QR code illustration"
                  className="feature-responsive-image"
                  loading="lazy"
                />
              </div>
            </div>
            <div className="feature-text-side">
              <span className="feature-kicker-tag">SHARE ANYWHERE</span>
              <h3 className="feature-headline">One link. Everyone can vote.</h3>
              <p className="feature-body-text">
                Share your poll with your audience using a simple public link. Works on every browser and mobile device with zero app installations, passwords, or barrier to entry.
              </p>
              <ul className="feature-check-list">
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>One-click copyable public voting URL</span>
                </li>
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Instant scannable QR code for slides and conference screens</span>
                </li>
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Dedicated distraction-free voter experience</span>
                </li>
              </ul>
              <div className="feature-cta-wrap">
                <Link to={isAuthenticated ? '/dashboard' : '/signup'}>
                  <Button variant="secondary" size="md" icon={<ShareIcon size={16} />}>
                    Explore Sharing
                  </Button>
                </Link>
              </div>
            </div>
          </div>

          {/* Feature 3: REAL-TIME RESULTS */}
          <div className="feature-block-row">
            <div className="feature-text-side">
              <span className="feature-kicker-tag">REAL-TIME RESULTS</span>
              <h3 className="feature-headline">Watch results happen live.</h3>
              <p className="feature-body-text">
                See votes update instantly without refreshing the page. Backed by high-throughput Redis Pub/Sub and WebSocket connections for sub-second synchronization.
              </p>
              <ul className="feature-check-list">
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>WebSocket + Redis Pub/Sub real-time event pipeline</span>
                </li>
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Smooth percentage transitions without screen flicker</span>
                </li>
                <li>
                  <CheckCircleIcon size={18} className="feature-list-check" />
                  <span>Presenter view optimized for live streaming and projector monitors</span>
                </li>
              </ul>
              <div className="feature-cta-wrap">
                <Link to={isAuthenticated ? '/dashboard' : '/login'}>
                  <Button variant="secondary" size="md" icon={<ChartIcon size={16} />}>
                    View Live Stream Demo
                  </Button>
                </Link>
              </div>
            </div>
            <div className="feature-visual-side">
              <div className="feature-image-card">
                <img
                  src="/images/live-results.png"
                  alt="Real-time results chart illustration"
                  className="feature-responsive-image"
                  loading="lazy"
                />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ====================================================================
          5. BOTTOM CALL TO ACTION
          ==================================================================== */}
      <section className="landing-cta-banner">
        <div className="cta-banner-card text-center">
          <span className="subtle-section-pill">GET STARTED IN SECONDS</span>
          <h2 className="cta-banner-title">Ready to launch your first live poll?</h2>
          <p className="cta-banner-subtitle">
            Join thousands of presenters and teams engaging audiences in real time.
          </p>
          <div className="cta-banner-buttons">
            <Link to={isAuthenticated ? '/create' : '/signup'}>
              <Button variant="primary" size="lg" icon={<PlusIcon size={18} />}>
                Create a Poll
              </Button>
            </Link>
            {!isAuthenticated && (
              <Link to="/login">
                <Button variant="outline" size="lg">
                  Sign In
                </Button>
              </Link>
            )}
          </div>
        </div>
      </section>
    </div>
  )
}
