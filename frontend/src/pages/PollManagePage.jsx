import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate, Link, useLocation } from 'react-router-dom'
import { pollsApi } from '../api/polls'
import { useLivePoll } from '../hooks/useLivePoll'
import { PollResultCard } from '../components/poll/PollResultCard'
import { Button } from '../components/common/Button'
import { Badge } from '../components/common/Badge'
import { CopyButton } from '../components/common/CopyButton'
import { QRCodeModal } from '../components/common/QRCodeModal'
import { Alert } from '../components/common/Alert'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { ChartIcon, VoteIcon, TrashIcon, ExternalLinkIcon, QrCodeIcon } from '../components/icons/Icons'
import { useAuth } from '../hooks/useAuth'
import { getPollShareUrl, getPollResultsUrl } from '../utils/url'

export function PollManagePage() {
  const { id, pollId } = useParams()
  const effectiveId = id || pollId

  const navigate = useNavigate()
  const location = useLocation()
  const { user } = useAuth()

  const [poll, setPoll] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [successMsg, setSuccessMsg] = useState(location.state?.message || null)
  const [statusUpdating, setStatusUpdating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [showQrModal, setShowQrModal] = useState(false)

  // Real-time live updates
  const { liveData, status: wsStatus, lastUpdated } = useLivePoll(effectiveId)

  const fetchPollDetails = useCallback(async () => {
    if (!effectiveId) return
    setLoading(true)
    setError(null)
    try {
      const data = await pollsApi.getPoll(effectiveId)
      setPoll(data)
    } catch (err) {
      setError(err.message || 'Failed to load poll details.')
    } finally {
      setLoading(false)
    }
  }, [effectiveId])

  useEffect(() => {
    fetchPollDetails()
  }, [fetchPollDetails])

  const handleToggleStatus = async () => {
    if (!poll || statusUpdating || !effectiveId) return
    setStatusUpdating(true)
    try {
      const nextStatus = !poll.isActive
      await pollsApi.setPollStatus(effectiveId, nextStatus)
      setPoll((prev) => ({ ...prev, isActive: nextStatus }))
      setSuccessMsg(`Poll status successfully updated to ${nextStatus ? 'Active (Open)' : 'Closed'}.`)
    } catch (err) {
      alert(`Failed to update poll status: ${err.message}`)
    } finally {
      setStatusUpdating(false)
    }
  }

  const handleDelete = async () => {
    if (!window.confirm('Are you sure you want to permanently delete this poll and all its recorded votes?')) {
      return
    }

    setDeleting(true)
    try {
      await pollsApi.deletePoll(effectiveId)
      navigate('/dashboard')
    } catch (err) {
      alert(`Failed to delete poll: ${err.message}`)
      setDeleting(false)
    }
  }

  if (loading && !poll) {
    return (
      <div className="page-center-container">
        <LoadingSpinner text="Loading poll management console..." size="lg" />
      </div>
    )
  }

  if (error || !poll) {
    return (
      <div className="page-center-container">
        <div className="error-panel glass-panel text-center max-w-md">
          <h2>Poll Not Found</h2>
          <p className="text-muted mt-8 mb-20">{error || 'This poll could not be located.'}</p>
          <Link to="/dashboard">
            <Button variant="primary" size="sm">
              Back to Dashboard
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  // Prevent non-creator from accessing management controls
  if (poll && user && poll.creatorId && user.id !== poll.creatorId) {
    return (
      <div className="page-center-container">
        <div className="error-panel glass-panel text-center max-w-md">
          <h2>Access Denied</h2>
          <p className="text-muted mt-8 mb-20">You do not have permission to manage this poll.</p>
          <Link to="/dashboard">
            <Button variant="primary" size="sm">
              Go to My Dashboard
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  const shareUrl = getPollShareUrl(effectiveId)
  const resultsUrl = getPollResultsUrl(effectiveId)

  // Combine live data with poll metadata
  const currentResults = liveData || {
    pollId: effectiveId,
    question: poll.question,
    isActive: poll.isActive,
    totalVotes: poll.options?.reduce((sum, o) => sum + (o.voteCount || 0), 0) || 0,
    results: poll.options?.map((o) => ({
      optionId: o.id,
      text: o.text,
      count: o.voteCount || 0,
      percentage: 0,
    })) || [],
  }

  return (
    <div className="manage-page-container">
      {/* Top action bar */}
      <div className="manage-header-row">
        <div>
          <div className="manage-title-group">
            <Link to="/dashboard" className="back-link">
              ← Dashboard
            </Link>
            <h1 className="page-title mt-4">Poll Management & Live Stream</h1>
          </div>
          <p className="page-subtitle">Configure poll access, share with audience, and monitor real-time votes</p>
        </div>

        <div className="manage-header-actions">
          <Button
            variant={poll.isActive ? 'outline' : 'primary'}
            size="md"
            onClick={handleToggleStatus}
            isLoading={statusUpdating}
          >
            {poll.isActive ? 'Close Voting' : 'Reopen Voting'}
          </Button>

          <Button
            variant="danger"
            size="md"
            onClick={handleDelete}
            isLoading={deleting}
            icon={<TrashIcon size={16} />}
          >
            Delete Poll
          </Button>
        </div>
      </div>

      {successMsg && (
        <Alert type="success" message={successMsg} onClose={() => setSuccessMsg(null)} className="mb-20" />
      )}

      {/* Shareable Link Banner */}
      <div className="share-banner glass-panel mb-24">
        <div className="share-banner-info">
          <span className="share-banner-label">Audience Voting URL (No login required)</span>
          <code className="share-url-box">{shareUrl}</code>
        </div>
        <div className="share-banner-actions">
          <Button
            variant="outline"
            size="md"
            icon={<QrCodeIcon size={16} />}
            onClick={() => setShowQrModal(true)}
          >
            Show QR Code
          </Button>
          <CopyButton text={shareUrl} label="Copy Vote Link" size="md" />
          <Link to={`/poll/${effectiveId}`} target="_blank" rel="noopener noreferrer">
            <Button variant="outline" size="md" icon={<VoteIcon size={16} />}>
              Test Vote Page <ExternalLinkIcon size={14} />
            </Button>
          </Link>
          <a href={resultsUrl} target="_blank" rel="noopener noreferrer">
            <Button variant="primary" size="md" icon={<ChartIcon size={16} />}>
              Full Screen Stream <ExternalLinkIcon size={14} />
            </Button>
          </a>
        </div>
      </div>

      {/* Live Results Preview with WebSocket updates */}
      <div className="manage-results-preview">
        <div className="section-title-row mb-16">
          <h3 className="text-lg font-bold">Real-Time Live Results Feed</h3>
          <Badge variant={wsStatus === 'connected' ? 'emerald' : 'amber'} isPulse={wsStatus === 'connected'}>
            {wsStatus === 'connected' ? 'WebSocket Active' : wsStatus}
          </Badge>
        </div>

        <PollResultCard
          results={currentResults}
          wsStatus={wsStatus}
          lastUpdated={lastUpdated}
          pollId={id}
          showShareActions={false}
        />
      </div>

      <QRCodeModal
        url={shareUrl}
        question={poll.question}
        isOpen={showQrModal}
        onClose={() => setShowQrModal(false)}
      />
    </div>
  )
}
