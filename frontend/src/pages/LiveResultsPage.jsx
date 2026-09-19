import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { pollsApi } from '../api/polls'
import { useLivePoll } from '../hooks/useLivePoll'
import { PollResultCard } from '../components/poll/PollResultCard'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { Button } from '../components/common/Button'
import { VoteIcon, ExternalLinkIcon, RadioIcon } from '../components/icons/Icons'

export function LiveResultsPage() {
  const { id } = useParams()
  const [initialPoll, setInitialPoll] = useState(null)
  const [initialLoading, setInitialLoading] = useState(true)
  const [initialError, setInitialError] = useState(null)
  const [showLogs, setShowLogs] = useState(false)

  // Real-time WebSocket hook
  const { liveData, status: wsStatus, lastUpdated, logs } = useLivePoll(id)

  // Fetch initial poll metadata once to ensure clean render if WebSocket takes a moment
  useEffect(() => {
    async function loadPoll() {
      setInitialLoading(true)
      try {
        const poll = await pollsApi.getPoll(id)
        setInitialPoll(poll)
      } catch (err) {
        setInitialError(err.message || 'Poll not found.')
      } finally {
        setInitialLoading(false)
      }
    }
    loadPoll()
  }, [id])

  if (initialLoading && !liveData) {
    return (
      <div className="page-center-container">
        <LoadingSpinner text="Connecting to Realtime WebSocket Stream..." size="lg" />
      </div>
    )
  }

  if (initialError && !liveData) {
    return (
      <div className="page-center-container">
        <div className="error-panel glass-panel text-center max-w-md">
          <h2>Poll Not Found</h2>
          <p className="text-muted mt-8 mb-20">{initialError}</p>
          <Link to="/">
            <Button variant="primary" size="sm">
              Return Home
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  // Combine initial poll metadata with live WebSocket results
  const resultsData = liveData || {
    pollId: id,
    question: initialPoll?.question || 'Live Poll',
    isActive: initialPoll?.isActive ?? true,
    totalVotes: initialPoll?.options?.reduce((sum, o) => sum + (o.voteCount || 0), 0) || 0,
    results: initialPoll?.options?.map((o) => ({
      optionId: o.id,
      text: o.text,
      count: o.voteCount || 0,
      percentage: 0,
    })) || [],
  }

  const voteUrl = `${window.location.origin}/polls/${id}`

  return (
    <div className="live-results-page-container">
      {/* Top Banner with live stream status */}
      <div className="live-stream-banner-row">
        <div className="live-channel-indicator">
          <RadioIcon size={18} />
          <span>Channel: <code>poll:events:{id.substring(0, 8)}...</code></span>
        </div>

        <div className="live-banner-actions">
          <button
            type="button"
            className="toggle-log-btn"
            onClick={() => setShowLogs(!showLogs)}
          >
            {showLogs ? 'Hide Stream Log' : `Show Stream Log (${logs.length})`}
          </button>

          <Link to={`/polls/${id}`} target="_blank" rel="noopener noreferrer" className="open-vote-btn">
            <VoteIcon size={16} />
            <span>Open Audience Vote Page</span>
            <ExternalLinkIcon size={14} />
          </Link>
        </div>
      </div>

      {/* Main Live Results Card */}
      <div className="live-results-wrapper">
        <PollResultCard
          results={resultsData}
          wsStatus={wsStatus}
          lastUpdated={lastUpdated}
          pollId={id}
          showShareActions={true}
        />
      </div>

      {/* Realtime WebSocket Stream Log Drawer */}
      {showLogs && (
        <div className="live-stream-log-drawer glass-panel">
          <div className="stream-log-header">
            <h4>📡 Realtime WebSocket Stream Activity (WS /api/polls/:id/live)</h4>
            <span className="log-count">{logs.length} events logged</span>
          </div>

          <div className="stream-log-body">
            {logs.length === 0 ? (
              <p className="no-logs-text">Waiting for votes... Cast a vote from another browser to see live events arrive!</p>
            ) : (
              logs.map((log) => (
                <div key={log.id} className="log-item-row">
                  <span className="log-time">[{log.time}]</span>
                  <span className="log-msg">{log.msg}</span>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
