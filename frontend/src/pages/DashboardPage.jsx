import { useState, useEffect, useCallback, useMemo } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { pollsApi } from '../api/polls'
import { useAuth } from '../hooks/useAuth'
import { Button } from '../components/common/Button'
import { Badge } from '../components/common/Badge'
import { QRCodeModal } from '../components/common/QRCodeModal'
import { Alert } from '../components/common/Alert'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { PollFlowLogo } from '../components/common/PollFlowLogo'
import { getPollShareUrl } from '../utils/url'
import {
  PlusIcon,
  ChartIcon,
  VoteIcon,
  TrashIcon,
  RefreshIcon,
  QrCodeIcon,
  LayoutDashboardIcon,
  ListIcon,
  SlidersIcon,
  SearchIcon,
  PlayIcon,
  TrendUpIcon,
  SparklesIcon,
  XIcon,
  CheckIcon,
  CopyIcon,
} from '../components/icons/Icons'

export function DashboardPage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const [polls, setPolls] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [deletingId, setDeletingId] = useState(null)
  const [togglingId, setTogglingId] = useState(null)
  const [activeQrPoll, setActiveQrPoll] = useState(null)
  const [currentNavSection, setCurrentNavSection] = useState('dashboard') // 'dashboard' | 'my-polls' | 'results' | 'settings' | 'help'
  const [filterTab, setFilterTab] = useState('all') // 'all' | 'active' | 'closed'
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState('newest') // 'newest' | 'votes' | 'title'
  const [copiedPollId, setCopiedPollId] = useState(null)

  const fetchPolls = useCallback(async () => {
    const token = localStorage.getItem('livepoll_jwt')
    if (!token) {
      setError('You are not currently logged in. Please sign in to view your dashboard.')
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)
    try {
      const data = await pollsApi.getMyPolls()
      setPolls(data)
    } catch (err) {
      if (err.status === 401) {
        setError('Your session has expired. Please sign in again.')
      } else if (err.message && err.message.toLowerCase().includes('network connection error')) {
        setError('Unable to reach the LivePoll backend server. Please ensure the backend is running.')
      } else {
        setError(err.message || 'Failed to load polls.')
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchPolls()
  }, [fetchPolls])

  const handleToggleStatus = async (pollId, currentStatus) => {
    setTogglingId(pollId)
    try {
      await pollsApi.setPollStatus(pollId, !currentStatus)
      setPolls((prev) =>
        prev.map((p) => (p.id === pollId ? { ...p, isActive: !currentStatus } : p))
      )
    } catch (err) {
      alert(`Failed to update poll status: ${err.message}`)
    } finally {
      setTogglingId(null)
    }
  }

  const handleDelete = async (pollId) => {
    if (!window.confirm('Are you sure you want to delete this poll? This action cannot be undone.')) {
      return
    }

    setDeletingId(pollId)
    try {
      await pollsApi.deletePoll(pollId)
      setPolls((prev) => prev.filter((p) => p.id !== pollId))
    } catch (err) {
      alert(`Failed to delete poll: ${err.message}`)
    } finally {
      setDeletingId(null)
    }
  }

  const handleCopyLink = async (pollId) => {
    const url = getPollShareUrl(pollId)
    try {
      await navigator.clipboard.writeText(url)
      setCopiedPollId(pollId)
      setTimeout(() => setCopiedPollId(null), 2400)
    } catch {
      alert(`Poll Link: ${url}`)
    }
  }

  // 13. DASHBOARD STATISTICS
  // Four compact statistics:
  // 1. Total Polls
  // 2. Active Polls
  // 3. Total Votes
  // 4. Responses Today
  const totalPolls = polls.length
  const activePolls = polls.filter((p) => p.isActive).length
  const totalVotesCast = polls.reduce(
    (acc, p) => acc + (p.options?.reduce((optAcc, o) => optAcc + (o.voteCount || 0), 0) || 0),
    0
  )

  // Calculate responses today (polls created today or estimate based on active responses)
  const todayStr = new Date().toDateString()
  const responsesToday = polls.reduce((acc, p) => {
    const pollDate = p.createdAt ? new Date(p.createdAt).toDateString() : null
    if (pollDate === todayStr) {
      return acc + (p.options?.reduce((optAcc, o) => optAcc + (o.voteCount || 0), 0) || 0)
    }
    return acc
  }, 0)

  // Filter & Sort logic
  const filteredAndSortedPolls = useMemo(() => {
    return polls
      .filter((p) => {
        // Tab filter
        if (filterTab === 'active' && !p.isActive) return false
        if (filterTab === 'closed' && p.isActive) return false

        // Search filter
        if (searchQuery.trim()) {
          const q = searchQuery.toLowerCase()
          const questionMatches = p.question?.toLowerCase().includes(q)
          const optionMatches = p.options?.some((o) => o.text?.toLowerCase().includes(q))
          if (!questionMatches && !optionMatches) return false
        }

        return true
      })
      .sort((a, b) => {
        if (sortBy === 'votes') {
          const aVotes = a.options?.reduce((s, o) => s + (o.voteCount || 0), 0) || 0
          const bVotes = b.options?.reduce((s, o) => s + (o.voteCount || 0), 0) || 0
          return bVotes - aVotes
        }
        if (sortBy === 'title') {
          return (a.question || '').localeCompare(b.question || '')
        }
        // default newest
        return new Date(b.createdAt || 0) - new Date(a.createdAt || 0)
      })
  }, [polls, filterTab, searchQuery, sortBy])

  const displayName = user?.name || user?.email?.split('@')[0] || 'Creator'

  return (
    <div className="saas-workspace-layout">
      {/* ====================================================================
          11. DASHBOARD SIDEBAR
          Sidebar:
          PollFlow logo
          Dashboard
          My Polls
          Create Poll
          Results
          Settings
          Help
          Logout
          ==================================================================== */}
      <aside className="saas-workspace-sidebar">
        <div className="sidebar-top-section">
          <Link to="/" className="sidebar-logo-link">
            <PollFlowLogo height={28} />
          </Link>

          <nav className="sidebar-nav-menu">
            <button
              type="button"
              className={`sidebar-nav-btn ${currentNavSection === 'dashboard' ? 'active' : ''}`}
              onClick={() => {
                setCurrentNavSection('dashboard')
                setFilterTab('all')
              }}
            >
              <LayoutDashboardIcon size={18} />
              <span>Dashboard</span>
            </button>

            <button
              type="button"
              className={`sidebar-nav-btn ${currentNavSection === 'my-polls' ? 'active' : ''}`}
              onClick={() => {
                setCurrentNavSection('my-polls')
                setFilterTab('all')
              }}
            >
              <ListIcon size={18} />
              <span>My Polls</span>
              {totalPolls > 0 && <span className="sidebar-counter-pill">{totalPolls}</span>}
            </button>

            <Link to="/create" className="sidebar-nav-btn create-action-highlight">
              <PlusIcon size={18} />
              <span>Create Poll</span>
            </Link>

            <button
              type="button"
              className={`sidebar-nav-btn ${currentNavSection === 'results' ? 'active' : ''}`}
              onClick={() => {
                setCurrentNavSection('results')
                setFilterTab('active')
              }}
            >
              <ChartIcon size={18} />
              <span>Results</span>
              {activePolls > 0 && <span className="sidebar-counter-pill live">{activePolls}</span>}
            </button>

            <button
              type="button"
              className={`sidebar-nav-btn ${currentNavSection === 'settings' ? 'active' : ''}`}
              onClick={() => setCurrentNavSection('settings')}
            >
              <SlidersIcon size={18} />
              <span>Settings</span>
            </button>

            <button
              type="button"
              className={`sidebar-nav-btn ${currentNavSection === 'help' ? 'active' : ''}`}
              onClick={() => setCurrentNavSection('help')}
            >
              <SparklesIcon size={18} />
              <span>Help & Docs</span>
            </button>
          </nav>

          <div className="sidebar-realtime-engine-pill">
            <div className="engine-status-row">
              <span className="live-pulsing-dot" />
              <span className="engine-status-text">Live Sync Engine</span>
            </div>
            <p className="engine-status-desc">WebSocket + Redis sub-second sync online</p>
          </div>
        </div>

        <div className="sidebar-bottom-section">
          <div className="sidebar-user-card">
            <div className="sidebar-avatar-circle">
              {displayName.charAt(0).toUpperCase()}
            </div>
            <div className="sidebar-user-details">
              <span className="sidebar-name-text">{displayName}</span>
              <span className="sidebar-email-text">{user?.email || 'Logged In'}</span>
            </div>
          </div>

          <button
            type="button"
            className="sidebar-logout-btn"
            onClick={() => {
              logout()
              navigate('/login')
            }}
          >
            <span>Log out</span>
          </button>
        </div>
      </aside>

      {/* ====================================================================
          MAIN CONTENT AREA
          ==================================================================== */}
      <main className="saas-workspace-main">
        {/* ==================================================================
            12. DASHBOARD HEADER
            Show:
            Dashboard
            "Manage your polls and see what's happening in real time."
            Primary CTA: Create Poll
            ================================================================== */}
        <header className="workspace-main-header">
          <div className="workspace-title-block">
            <h1 className="workspace-title">Dashboard</h1>
            <p className="workspace-subtitle">
              Manage your polls and see what's happening in real time.
            </p>
          </div>

          <div className="workspace-header-actions">
            <Button
              variant="secondary"
              size="sm"
              onClick={fetchPolls}
              disabled={loading}
              icon={<RefreshIcon size={14} />}
            >
              Refresh
            </Button>

            <Link to="/create">
              <Button variant="primary" size="md" icon={<PlusIcon size={16} />}>
                Create Poll
              </Button>
            </Link>
          </div>
        </header>

        {/* ==================================================================
            13. DASHBOARD STATISTICS (Four compact statistics)
            Total Polls | Active Polls | Total Votes | Responses Today
            ================================================================== */}
        <section className="compact-statistics-grid">
          <div className="compact-stat-card">
            <div className="stat-card-top">
              <span className="stat-metric-title">Total Polls</span>
              <LayoutDashboardIcon size={16} className="stat-metric-icon" />
            </div>
            <div className="stat-card-value">{totalPolls}</div>
            <span className="stat-metric-note">Lifetime questions created</span>
          </div>

          <div className="compact-stat-card active-card">
            <div className="stat-card-top">
              <span className="stat-metric-title">Active Polls</span>
              <span className="live-dot-green" />
            </div>
            <div className="stat-card-value emerald-text">{activePolls}</div>
            <span className="stat-metric-note">Accepting live responses</span>
          </div>

          <div className="compact-stat-card">
            <div className="stat-card-top">
              <span className="stat-metric-title">Total Votes</span>
              <TrendUpIcon size={16} className="stat-metric-icon" />
            </div>
            <div className="stat-card-value blue-text">{totalVotesCast}</div>
            <span className="stat-metric-note">Submissions recorded live</span>
          </div>

          <div className="compact-stat-card">
            <div className="stat-card-top">
              <span className="stat-metric-title">Responses Today</span>
              <ChartIcon size={16} className="stat-metric-icon" />
            </div>
            <div className="stat-card-value">{responsesToday}</div>
            <span className="stat-metric-note">Votes submitted today</span>
          </div>
        </section>

        {/* Error notification if any */}
        {error && (
          <div className="mb-20">
            <Alert type="error" message={error} onClose={() => setError(null)} />
          </div>
        )}

        {/* ==================================================================
            SETTINGS OR HELP VIEWS (When selected from sidebar)
            ================================================================== */}
        {currentNavSection === 'settings' && (
          <div className="dashboard-subview-panel">
            <div className="subview-header">
              <h2>Account & Workspace Settings</h2>
              <p>Manage your live polling defaults and profile preferences.</p>
            </div>
            <div className="subview-card">
              <div className="subview-item">
                <span className="subview-label">Creator Email</span>
                <span className="subview-value">{user?.email}</span>
              </div>
              <div className="subview-item">
                <span className="subview-label">Realtime Engine</span>
                <span className="subview-value">WebSocket via Go & Redis</span>
              </div>
              <div className="subview-item">
                <span className="subview-label">Default Response Limit</span>
                <span className="subview-value">Single vote per participant</span>
              </div>
            </div>
          </div>
        )}

        {currentNavSection === 'help' && (
          <div className="dashboard-subview-panel">
            <div className="subview-header">
              <h2>Help & Audience Engagement Guide</h2>
              <p>Best practices for gathering live responses in webinars, slides, and classrooms.</p>
            </div>
            <div className="subview-card">
              <h4>Tips for High Response Rates:</h4>
              <ul className="help-tips-list">
                <li>Project the instant QR code onto your opening slide before your talk begins.</li>
                <li>Keep questions short and options between 2 to 4 choices for fast decision-making.</li>
                <li>Open the Live Results view in a separate monitor window to present live bar animations.</li>
              </ul>
            </div>
          </div>
        )}

        {/* ==================================================================
            14. ACTIVE POLLS & 15. MY POLLS TABLE / LIST
            Columns: Poll | Status | Votes | Created | Actions
            Actions: View | Share | Results | Manage
            Mobile: Stacked cards (zero horizontal scroll)
            ================================================================== */}
        {(currentNavSection === 'dashboard' ||
          currentNavSection === 'my-polls' ||
          currentNavSection === 'results') && (
          <section className="polls-table-section">
            {/* Control & Filter Toolbar */}
            <div className="polls-toolbar-row">
              {/* Filter Tabs */}
              <div className="filter-tab-pill-group">
                <button
                  type="button"
                  className={`filter-tab-pill ${filterTab === 'all' ? 'active' : ''}`}
                  onClick={() => setFilterTab('all')}
                >
                  All Polls <span className="tab-count">{polls.length}</span>
                </button>
                <button
                  type="button"
                  className={`filter-tab-pill ${filterTab === 'active' ? 'active' : ''}`}
                  onClick={() => setFilterTab('active')}
                >
                  Active <span className="tab-count green">{activePolls}</span>
                </button>
                <button
                  type="button"
                  className={`filter-tab-pill ${filterTab === 'closed' ? 'active' : ''}`}
                  onClick={() => setFilterTab('closed')}
                >
                  Closed <span className="tab-count">{polls.length - activePolls}</span>
                </button>
              </div>

              {/* Search & Sort Controls */}
              <div className="toolbar-search-sort-cluster">
                <div className="search-field-wrapper">
                  <SearchIcon size={14} className="search-field-icon" />
                  <input
                    type="text"
                    placeholder="Search polls..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="search-field-input"
                  />
                  {searchQuery && (
                    <button
                      type="button"
                      className="search-clear-btn"
                      onClick={() => setSearchQuery('')}
                      title="Clear search"
                    >
                      <XIcon size={14} />
                    </button>
                  )}
                </div>

                <div className="sort-select-wrapper">
                  <select
                    value={sortBy}
                    onChange={(e) => setSortBy(e.target.value)}
                    className="sort-select-element"
                    aria-label="Sort polls"
                  >
                    <option value="newest">Newest First</option>
                    <option value="votes">Most Votes</option>
                    <option value="title">Question (A-Z)</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Table / List Body */}
            {loading ? (
              <div className="polls-loading-card">
                <LoadingSpinner text="Loading your polls..." size="md" />
              </div>
            ) : filteredAndSortedPolls.length === 0 ? (
              <div className="polls-empty-card">
                {searchQuery ? (
                  <>
                    <div className="empty-icon-bubble">
                      <SearchIcon size={22} />
                    </div>
                    <h3>No matching polls found</h3>
                    <p>No polls matched your search query "{searchQuery}".</p>
                    <Button variant="secondary" size="sm" onClick={() => setSearchQuery('')}>
                      Clear Search
                    </Button>
                  </>
                ) : (
                  <>
                    <div className="empty-icon-bubble">
                      <SparklesIcon size={24} />
                    </div>
                    <h3>No polls created yet</h3>
                    <p>
                      Ask your first question and start collecting real-time responses from your audience.
                    </p>
                    <Link to="/create" className="mt-16">
                      <Button variant="primary" size="md" icon={<PlusIcon size={16} />}>
                        Create Your First Poll
                      </Button>
                    </Link>
                  </>
                )}
              </div>
            ) : (
              <>
                {/* 15. DESKTOP VIEW: Professional Table */}
                <div className="desktop-polls-table-container">
                  <table className="saas-polls-table">
                    <thead>
                      <tr>
                        <th className="th-poll">Poll</th>
                        <th className="th-status">Status</th>
                        <th className="th-votes">Votes</th>
                        <th className="th-created">Created</th>
                        <th className="th-actions text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredAndSortedPolls.map((poll) => {
                        const pollVotes =
                          poll.options?.reduce((sum, o) => sum + (o.voteCount || 0), 0) || 0
                        const shareUrl = getPollShareUrl(poll.id)
                        const formattedDate = poll.createdAt
                          ? new Date(poll.createdAt).toLocaleDateString(undefined, {
                              month: 'short',
                              day: 'numeric',
                              year: 'numeric',
                            })
                          : 'Recent'

                        const isCopied = copiedPollId === poll.id

                        return (
                          <tr key={poll.id} className="table-poll-row">
                            {/* Column: Poll Question & Options count */}
                            <td className="td-poll">
                              <div className="poll-title-cell">
                                <Link
                                  to={`/polls/${poll.id}/manage`}
                                  className="poll-question-link"
                                  title="Manage poll"
                                >
                                  {poll.question}
                                </Link>
                                <span className="poll-options-subtext">
                                  {poll.options?.length || 0} answer options
                                </span>
                              </div>
                            </td>

                            {/* Column: Status */}
                            <td className="td-status">
                              <Badge variant={poll.isActive ? 'active' : 'closed'} size="sm">
                                {poll.isActive && <span className="live-dot-pulse-tiny" />}
                                {poll.isActive ? 'Active' : 'Closed'}
                              </Badge>
                            </td>

                            {/* Column: Votes */}
                            <td className="td-votes">
                              <span className="votes-count-pill">
                                <strong>{pollVotes}</strong> {pollVotes === 1 ? 'vote' : 'votes'}
                              </span>
                            </td>

                            {/* Column: Created */}
                            <td className="td-created">
                              <span className="date-cell-text">{formattedDate}</span>
                            </td>

                            {/* Column: Actions (View, Share, Results, Manage) */}
                            <td className="td-actions text-right">
                              <div className="table-actions-cluster">
                                {/* View (Audience Page) */}
                                <Link
                                  to={`/polls/${poll.id}`}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="table-action-btn"
                                  title="View audience voting page in new tab"
                                >
                                  <VoteIcon size={14} />
                                  <span>View</span>
                                </Link>

                                {/* Share (Copy Link) */}
                                <button
                                  type="button"
                                  className={`table-action-btn ${isCopied ? 'copied' : ''}`}
                                  onClick={() => handleCopyLink(poll.id)}
                                  title="Copy voting link"
                                >
                                  {isCopied ? <CheckIcon size={14} /> : <CopyIcon size={14} />}
                                  <span>{isCopied ? 'Copied' : 'Share'}</span>
                                </button>

                                {/* QR Code modal */}
                                <button
                                  type="button"
                                  className="table-action-btn"
                                  onClick={() =>
                                    setActiveQrPoll({ url: shareUrl, question: poll.question })
                                  }
                                  title="Show QR code"
                                >
                                  <QrCodeIcon size={14} />
                                </button>

                                {/* Results (Live Presenter Screen) */}
                                <Link
                                  to={`/polls/${poll.id}/results`}
                                  className="table-action-btn highlight-results"
                                  title="Open live results presenter screen"
                                >
                                  <PlayIcon size={13} />
                                  <span>Results</span>
                                </Link>

                                {/* Toggle Status (Close / Reopen) */}
                                <button
                                  type="button"
                                  className="table-action-btn"
                                  onClick={() => handleToggleStatus(poll)}
                                  disabled={togglingId === poll.id}
                                  title={poll.isActive ? 'Close poll' : 'Reopen poll'}
                                >
                                  <span>{poll.isActive ? 'Close' : 'Reopen'}</span>
                                </button>

                                {/* Manage */}
                                <Link
                                  to={`/polls/${poll.id}/manage`}
                                  className="table-action-btn"
                                  title="Configure poll settings"
                                >
                                  <SlidersIcon size={14} />
                                  <span>Manage</span>
                                </Link>

                                {/* Delete */}
                                <button
                                  type="button"
                                  className="table-action-btn danger-trash"
                                  onClick={() => handleDelete(poll.id)}
                                  disabled={deletingId === poll.id}
                                  title="Delete poll"
                                >
                                  <TrashIcon size={14} />
                                </button>
                              </div>
                            </td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </table>
                </div>

                {/* 15. MOBILE VIEW: Stacked Cards (No horizontal scroll!) */}
                <div className="mobile-polls-cards-stack">
                  {filteredAndSortedPolls.map((poll) => {
                    const pollVotes =
                      poll.options?.reduce((sum, o) => sum + (o.voteCount || 0), 0) || 0
                    const shareUrl = getPollShareUrl(poll.id)
                    const formattedDate = poll.createdAt
                      ? new Date(poll.createdAt).toLocaleDateString(undefined, {
                          month: 'short',
                          day: 'numeric',
                        })
                      : 'Recent'
                    const isCopied = copiedPollId === poll.id

                    return (
                      <div key={poll.id} className="mobile-poll-card">
                        <div className="mobile-card-top-row">
                          <Badge variant={poll.isActive ? 'active' : 'closed'} size="sm">
                            {poll.isActive && <span className="live-dot-pulse-tiny" />}
                            {poll.isActive ? 'Active' : 'Closed'}
                          </Badge>
                          <span className="mobile-date-text">{formattedDate}</span>
                        </div>

                        <h3 className="mobile-poll-title">
                          <Link to={`/polls/${poll.id}/manage`}>{poll.question}</Link>
                        </h3>

                        <div className="mobile-card-meta-row">
                          <span className="mobile-vote-count">
                            <strong>{pollVotes}</strong> votes
                          </span>
                          <span className="dot-sep">·</span>
                          <span className="mobile-options-count">
                            {poll.options?.length || 0} options
                          </span>
                        </div>

                        {/* Mobile Actions Cluster */}
                        <div className="mobile-card-actions-grid">
                          <Link
                            to={`/polls/${poll.id}/results`}
                            className="mobile-action-link primary"
                          >
                            <PlayIcon size={14} />
                            <span>Results</span>
                          </Link>

                          <button
                            type="button"
                            className={`mobile-action-link ${isCopied ? 'copied' : ''}`}
                            onClick={() => handleCopyLink(poll.id)}
                          >
                            {isCopied ? <CheckIcon size={14} /> : <CopyIcon size={14} />}
                            <span>{isCopied ? 'Copied' : 'Share'}</span>
                          </button>

                          <button
                            type="button"
                            className="mobile-action-link"
                            onClick={() =>
                              setActiveQrPoll({ url: shareUrl, question: poll.question })
                            }
                            title="Show QR code"
                          >
                            <span>QR Code</span>
                          </button>

                          <Link to={`/polls/${poll.id}`} target="_blank" rel="noreferrer" className="mobile-action-link">
                            <VoteIcon size={14} />
                            <span>View</span>
                          </Link>

                          <Link to={`/polls/${poll.id}/manage`} className="mobile-action-link">
                            <SlidersIcon size={14} />
                            <span>Manage</span>
                          </Link>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </>
            )}
          </section>
        )}
      </main>

      {/* QR Code Modal for quick sharing */}
      {activeQrPoll && (
        <QRCodeModal
          url={activeQrPoll.url}
          question={activeQrPoll.question}
          isOpen={!!activeQrPoll}
          onClose={() => setActiveQrPoll(null)}
        />
      )}
    </div>
  )
}
