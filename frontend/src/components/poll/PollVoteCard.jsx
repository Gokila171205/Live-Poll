import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../common/Button'
import { Alert } from '../components/../common/Alert'
import { Badge } from '../common/Badge'
import { PollFlowLogo } from '../common/PollFlowLogo'
import { CheckIcon, ChartIcon } from '../icons/Icons'

export function PollVoteCard({
  poll,
  onVote,
  isVoting = false,
  hasVoted = false,
  votedOptionId = null,
  error = null,
}) {
  const [selectedOptionId, setSelectedOptionId] = useState(votedOptionId)

  const handleSelect = (optionId) => {
    if (hasVoted || !poll.isActive) return
    setSelectedOptionId(optionId)
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!selectedOptionId || hasVoted || isVoting || !poll.isActive) return
    onVote(selectedOptionId)
  }

  // Calculate live results for post-vote display
  const totalVotes =
    poll.options?.reduce((sum, o) => sum + (o.voteCount || 0), 0) || 0

  return (
    <div className="voter-experience-card">
      {/* 16. TOP: PollFlow Logo & Live status */}
      <div className="voter-card-header">
        <Link to="/" className="voter-logo-link">
          <PollFlowLogo height={24} showText={true} />
        </Link>
        {poll.isActive ? (
          <span className="voter-live-badge">
            <span className="live-dot-green" />
            LIVE POLL
          </span>
        ) : (
          <span className="voter-closed-badge">CLOSED</span>
        )}
      </div>

      {/* CENTER: Poll Question */}
      <div className="voter-question-block">
        <h1 className="voter-headline-question">{poll.question}</h1>
        <p className="voter-subtext-prompt">
          {hasVoted
            ? 'Thank you for participating! Here are the current results:'
            : !poll.isActive
            ? 'This poll is no longer accepting new responses.'
            : 'Select one option below and submit your vote:'}
        </p>
      </div>

      {!poll.isActive && !hasVoted && (
        <Alert
          type="warning"
          message="Voting for this poll has ended."
          className="mb-16"
        />
      )}

      {error && <Alert type="error" message={error} className="mb-16" />}

      {/* 16. AFTER VOTING: "✓ Vote submitted" */}
      {hasVoted && (
        <div className="voter-success-banner">
          <div className="voter-success-icon-wrap">
            <CheckIcon size={18} />
          </div>
          <div className="voter-success-text-col">
            <span className="voter-success-title">✓ Vote submitted</span>
            <span className="voter-success-sub">Your response has been counted in real time.</span>
          </div>
        </div>
      )}

      {/* BEFORE VOTING: Answer options form */}
      {!hasVoted ? (
        <form onSubmit={handleSubmit} className="voter-options-form">
          <div className="voter-options-stack">
            {poll.options?.map((opt) => {
              const isSelected = selectedOptionId === opt.id

              return (
                <div
                  key={opt.id}
                  className={`voter-option-row ${isSelected ? 'selected' : ''} ${!poll.isActive ? 'disabled' : ''}`}
                  onClick={() => handleSelect(opt.id)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') handleSelect(opt.id)
                  }}
                >
                  <div className="voter-radio-wrap">
                    <span className={`custom-radio-circle ${isSelected ? 'checked' : ''}`}>
                      {isSelected && <span className="custom-radio-inner" />}
                    </span>
                  </div>
                  <span className="voter-option-label">{opt.text}</span>
                </div>
              )
            })}
          </div>

          <div className="voter-cta-wrap mt-24">
            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isVoting}
              disabled={!selectedOptionId || !poll.isActive}
              className="w-full voter-submit-btn"
            >
              {!poll.isActive ? 'Voting Closed' : 'Submit Vote'}
            </Button>
          </div>
        </form>
      ) : (
        /* AFTER VOTING: Live Results Bars */
        <div className="voter-results-section mt-16">
          <div className="voter-results-summary-row">
            <span className="voter-results-label">Live Outcome</span>
            <span className="voter-results-total">
              <strong>{totalVotes}</strong> total {totalVotes === 1 ? 'vote' : 'votes'}
            </span>
          </div>

          <div className="voter-results-bars-list">
            {poll.options?.map((opt) => {
              const votes = opt.voteCount || 0
              const pct = totalVotes > 0 ? Math.round((votes / totalVotes) * 100) : 0
              const isUserChoice = votedOptionId === opt.id

              return (
                <div
                  key={opt.id}
                  className={`voter-result-bar-item ${isUserChoice ? 'user-choice' : ''}`}
                >
                  <div className="voter-result-meta-row">
                    <div className="voter-result-left">
                      <span className="voter-result-opt-name">{opt.text}</span>
                      {isUserChoice && (
                        <span className="voter-your-vote-tag">Your choice</span>
                      )}
                    </div>
                    <div className="voter-result-right">
                      <span className="voter-result-pct">{pct}%</span>
                      <span className="voter-result-count">({votes})</span>
                    </div>
                  </div>

                  <div className="voter-progress-track">
                    <div
                      className={`voter-progress-fill ${isUserChoice ? 'user-fill' : ''}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>

          <div className="voter-fullscreen-link-wrap">
            <Link to={`/polls/${poll.id}/results`} className="voter-live-screen-link">
              <ChartIcon size={14} />
              <span>Open live presenter screen</span>
            </Link>
          </div>
        </div>
      )}

      {/* FOOTER: Strict isolation from creator controls */}
      <div className="voter-card-footer">
        <span className="voter-powered-by">Powered by PollFlow</span>
      </div>
    </div>
  )
}
