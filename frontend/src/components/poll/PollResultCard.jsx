import { useState } from 'react'
import { Link } from 'react-router-dom'
import { CopyButton } from '../common/CopyButton'
import { Button } from '../common/Button'
import { QRCodeModal } from '../common/QRCodeModal'
import { VoteIcon, ExternalLinkIcon, QrCodeIcon } from '../icons/Icons'
import { PollFlowLogo } from '../common/PollFlowLogo'
import { getPollShareUrl } from '../../utils/url'

export function PollResultCard({
  results,
  wsStatus = 'connected',
  lastUpdated = null,
  pollId,
  showShareActions = true,
}) {
  const [showQrModal, setShowQrModal] = useState(false)

  if (!results) return null

  const targetPollId = pollId || results.pollId
  const shareUrl = getPollShareUrl(targetPollId)
  const totalVotes = results.totalVotes || 0

  // Calculate highest vote count to highlight leading option
  const rawList = results.results || results.options || []
  const maxCount = Math.max(...rawList.map((o) => o.count ?? o.voteCount ?? 0), 0)

  const optionList = rawList.map((opt) => {
    const count = opt.count ?? opt.voteCount ?? 0
    const pct = opt.percentage ?? (totalVotes > 0 ? ((count / totalVotes) * 100).toFixed(1) : 0)
    const isLeader = maxCount > 0 && count === maxCount

    return {
      id: opt.optionId || opt.id,
      text: opt.text,
      count,
      percentage: typeof pct === 'number' ? pct.toFixed(0) : pct,
      isLeader,
    }
  })

  return (
    <>
      <div className="live-results-presentation-card">
        {/* Top Branding & Connection Bar */}
        <div className="presentation-card-top-nav">
          <Link to="/" className="presentation-brand-link">
            <PollFlowLogo height={24} />
          </Link>

          {showShareActions && (
            <div className="presentation-actions-right">
              <Button
                variant="outline"
                size="sm"
                icon={<QrCodeIcon size={15} />}
                onClick={() => setShowQrModal(true)}
              >
                QR Code
              </Button>
              <CopyButton text={shareUrl} label="Copy Vote Link" size="sm" />
              <a
                href={shareUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="presentation-voter-link"
              >
                <VoteIcon size={14} />
                <span>Audience Vote</span>
                <ExternalLinkIcon size={12} />
              </a>
            </div>
          )}
        </div>

        {/* 17. HEADER: LIVE ●, Poll question, Total Votes */}
        <div className="presentation-header-block">
          <div className="presentation-live-status-row">
            <span className="live-stream-badge">
              <span className="live-dot-red" />
              LIVE ●
            </span>
            {results.isActive === false && (
              <span className="presentation-closed-tag">VOTING CLOSED</span>
            )}
            <span className="presentation-engine-tag">
              {wsStatus === 'connected' ? 'WebSocket Sync Active' : 'Connecting to Stream...'}
            </span>
          </div>

          <h1 className="presentation-question-title">{results.question}</h1>

          <div className="presentation-vote-tally-strip">
            <span className="total-votes-number">{totalVotes}</span>
            <span className="total-votes-label">
              {totalVotes === 1 ? 'Total Vote' : 'Total Votes'}
            </span>
            {lastUpdated && (
              <span className="last-sync-time">
                · Last update {lastUpdated.toLocaleTimeString()}
              </span>
            )}
          </div>
        </div>

        {/* 17. RESULT BARS (Animated smoothly on WebSocket messages) */}
        <div className="presentation-bars-stack">
          {optionList.map((opt) => (
            <div
              key={opt.id}
              className={`presentation-bar-row ${opt.isLeader && totalVotes > 0 ? 'leader-row' : ''}`}
            >
              <div className="presentation-bar-meta-line">
                <div className="meta-line-left">
                  <span className="presentation-opt-title">{opt.text}</span>
                  {opt.isLeader && totalVotes > 0 && (
                    <span className="leading-indicator-badge">★ Leader</span>
                  )}
                </div>
                <div className="meta-line-right">
                  <span className="presentation-pct-val">{opt.percentage}%</span>
                  <span className="presentation-count-val">
                    {opt.count} {opt.count === 1 ? 'vote' : 'votes'}
                  </span>
                </div>
              </div>

              {/* Progress Bar Track & Animated Fill */}
              <div className="presentation-track">
                <div
                  className={`presentation-fill ${opt.isLeader && totalVotes > 0 ? 'leader-fill' : ''}`}
                  style={{
                    width: `${Math.min(Math.max(Number(opt.percentage), 0), 100)}%`,
                  }}
                />
              </div>
            </div>
          ))}
        </div>

        {/* Empty state if 0 votes yet */}
        {totalVotes === 0 && (
          <div className="presentation-waiting-banner">
            <span className="waiting-pulse-dot" />
            <p>
              Waiting for first responses... Share the voting link or display the QR code!
            </p>
          </div>
        )}

        <div className="presentation-card-bottom-bar">
          <span className="engine-credits">Powered by PollFlow Real-Time Architecture</span>
          <span className="audience-help-text">Audience can vote live at: <code>{shareUrl}</code></span>
        </div>
      </div>

      {/* QR Code Sharing Modal */}
      <QRCodeModal
        url={shareUrl}
        question={results.question}
        isOpen={showQrModal}
        onClose={() => setShowQrModal(false)}
      />
    </>
  )
}
