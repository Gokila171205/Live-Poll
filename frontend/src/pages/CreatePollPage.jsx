import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { pollsApi } from '../api/polls'
import { Button } from '../components/common/Button'
import { Alert } from '../components/common/Alert'
import { Toggle } from '../components/common/Toggle'
import { QRCodeModal } from '../components/common/QRCodeModal'
import {
  PlusIcon,
  XIcon,
  CheckIcon,
  CopyIcon,
  ExternalLinkIcon,
  ChartIcon,
  QrCodeIcon,
  EyeIcon,
  VoteIcon,
} from '../components/icons/Icons'
import { PollFlowLogo } from '../components/common/PollFlowLogo'
import { getPollShareUrl, getPollResultsUrl } from '../utils/url'

export function CreatePollPage() {
  const navigate = useNavigate()

  // Form State - initialized cleanly with empty strings for controlled inputs
  const [question, setQuestion] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [oneVotePerParticipant, setOneVotePerParticipant] = useState(true)
  const [showResultsAfterVoting, setShowResultsAfterVoting] = useState(true)
  const [initialStatus, setInitialStatus] = useState('active') // 'active' | 'closed'

  // UI & Lifecycle State
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(false)
  const [createdPollData, setCreatedPollData] = useState(null) // Holds created poll info for Success State
  const [copiedLink, setCopiedLink] = useState(false)
  const [showQrModal, setShowQrModal] = useState(false)
  const [previewSelectedIdx, setPreviewSelectedIdx] = useState(null)

  const handleAddOption = () => {
    if (options.length >= 10) return
    setOptions([...options, ''])
  }

  const handleRemoveOption = (index) => {
    if (options.length <= 2) return
    setOptions(options.filter((_, idx) => idx !== index))
  }

  const handleOptionChange = (index, value) => {
    const updated = [...options]
    updated[index] = value
    setOptions(updated)
    if (error) setError(null)
  }

  // Validate form inputs
  const validateForm = () => {
    const trimmedQuestion = question.trim()
    if (!trimmedQuestion) {
      return 'Please enter a poll question.'
    }
    if (trimmedQuestion.length < 3) {
      return 'Question must be at least 3 characters long.'
    }

    const trimmedOptions = options.map((o) => o.trim())
    for (let i = 0; i < trimmedOptions.length; i++) {
      if (!trimmedOptions[i]) {
        return `Option ${i + 1} cannot be empty.`
      }
    }

    if (trimmedOptions.length < 2) {
      return 'Please provide at least 2 options.'
    }

    // Duplicate options check (case-insensitive)
    const seen = new Set()
    for (const opt of trimmedOptions) {
      const lower = opt.toLowerCase()
      if (seen.has(lower)) {
        return `Duplicate option "${opt}" detected. All options must be unique.`
      }
      seen.add(lower)
    }

    return null
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)

    const validationErr = validateForm()
    if (validationErr) {
      setError(validationErr)
      return
    }

    const trimmedQuestion = question.trim()
    const cleanedOptions = options.map((o) => o.trim())

    setLoading(true)
    try {
      const createdPoll = await pollsApi.createPoll(trimmedQuestion, cleanedOptions)

      // If creator configured status as closed, set it
      if (initialStatus === 'closed' && createdPoll.id) {
        try {
          await pollsApi.setPollStatus(createdPoll.id, false)
          createdPoll.isActive = false
        } catch {
          // ignore non-fatal
        }
      }

      // Transition to polished SUCCESS EXPERIENCE rather than abrupt redirect
      setCreatedPollData(createdPoll)
    } catch (err) {
      setError(err.message || 'Failed to create poll. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleCopyLink = async (url) => {
    try {
      await navigator.clipboard.writeText(url)
      setCopiedLink(true)
      setTimeout(() => setCopiedLink(false), 2200)
    } catch {
      alert(`Poll Link: ${url}`)
    }
  }

  const scrollToPreview = () => {
    const el = document.getElementById('live-preview-section')
    if (el) el.scrollIntoView({ behavior: 'smooth' })
  }

  // ====================================================================
  // 19. SUCCESS EXPERIENCE
  // Show polished success state after poll creation
  // ====================================================================
  if (createdPollData) {
    const pollShareUrl = getPollShareUrl(createdPollData.id)
    const resultsUrl = getPollResultsUrl(createdPollData.id)

    return (
      <div className="create-success-page-shell">
        <div className="create-success-card">
          <div className="success-icon-badge">
            <CheckIcon size={28} />
          </div>

          <span className="success-kicker">✓ POLL CREATED SUCCESSFULLY</span>
          <h1 className="success-heading">Your poll is ready to share with your audience!</h1>
          <p className="success-subheading">
            "{createdPollData.question || question}"
          </p>

          <p className="success-hint-text">
            Anyone with this link can vote. Share it in your presentation, webinar chat, or slides.
          </p>

          {/* Share Link Box */}
          <div className="success-link-box">
            <span className="success-link-url">{pollShareUrl}</span>
            <button
              type="button"
              className={`success-copy-btn ${copiedLink ? 'copied' : ''}`}
              onClick={() => handleCopyLink(pollShareUrl)}
            >
              {copiedLink ? <CheckIcon size={16} /> : <CopyIcon size={16} />}
              <span>{copiedLink ? 'Copied!' : 'Copy Link'}</span>
            </button>
          </div>

          {/* Action CTAs */}
          <div className="success-action-buttons">
            <a
              href={pollShareUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="success-btn-action"
            >
              <Button variant="secondary" size="md" icon={<VoteIcon size={16} />}>
                Open Poll <ExternalLinkIcon size={13} />
              </Button>
            </a>

            <a
              href={resultsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="success-btn-action"
            >
              <Button variant="primary" size="md" icon={<ChartIcon size={16} />}>
                View Live Results
              </Button>
            </a>

            <Button
              variant="outline"
              size="md"
              icon={<QrCodeIcon size={16} />}
              onClick={() => setShowQrModal(true)}
            >
              Show QR Code
            </Button>
          </div>

          <div className="success-card-footer">
            <Link to="/dashboard" className="success-dashboard-link">
              Go to Dashboard
            </Link>
            <span className="dot-divider">·</span>
            <button
              type="button"
              className="success-another-btn"
              onClick={() => {
                setCreatedPollData(null)
                setQuestion('')
                setOptions(['', ''])
                setError(null)
              }}
            >
              Create Another Poll
            </button>
          </div>
        </div>

        {showQrModal && (
          <QRCodeModal
            url={pollShareUrl}
            question={createdPollData.question || question}
            isOpen={showQrModal}
            onClose={() => setShowQrModal(false)}
          />
        )}
      </div>
    )
  }

  // ====================================================================
  // 9. CREATE POLL PAGE (Two-Column Survey Builder: Left Builder, Right Preview)
  // ====================================================================
  return (
    <div className="create-poll-workspace">
      <div className="create-poll-header-strip">
        <div className="header-strip-container">
          <div className="header-text-block">
            <h1 className="builder-main-title">Create a new poll</h1>
            <p className="builder-sub-title">
              Ask a question and get real-time responses from your audience.
            </p>
          </div>
          <div className="header-badge-block">
            <span className="builder-live-tag">
              <span className="live-dot-blue" />
              Live Survey Builder
            </span>
          </div>
        </div>
      </div>

      <div className="builder-two-column-layout">
        {/* ==================================================================
            LEFT COLUMN: POLL BUILDER FORM
            ================================================================== */}
        <div className="builder-form-column">
          <div className="builder-card">
            {error && (
              <Alert
                type="error"
                message={error}
                onClose={() => setError(null)}
                className="mb-20"
              />
            )}

            <form onSubmit={handleSubmit} className="poll-form">
              {/* SECTION: QUESTION */}
              <div className="form-section-block">
                <div className="section-label-row">
                  <label htmlFor="poll-question-input" className="builder-field-label">
                    Poll question <span className="required-asterisk">*</span>
                  </label>
                  <span className="char-counter">{question.length} / 200</span>
                </div>

                <input
                  id="poll-question-input"
                  type="text"
                  className="builder-question-input"
                  placeholder="What would you like to ask your audience?"
                  value={question}
                  onChange={(e) => {
                    setQuestion(e.target.value)
                    if (error) setError(null)
                  }}
                  maxLength={200}
                  required
                />
                <span className="builder-helper-text">
                  Make your question concise and direct for faster responses.
                </span>
              </div>

              {/* SECTION: OPTIONS */}
              <div className="form-section-block mt-24">
                <div className="section-label-row">
                  <label className="builder-field-label">
                    Answer options <span className="required-asterisk">*</span>
                  </label>
                  <span className="options-count-badge">{options.length} options</span>
                </div>

                <div className="options-builder-stack">
                  {options.map((opt, idx) => (
                    <div key={idx} className="builder-option-row">
                      <span className="builder-option-num">Option {idx + 1}</span>
                      <div className="builder-option-input-wrap">
                        <input
                          type="text"
                          className="builder-option-input"
                          placeholder={`Enter option ${idx + 1}`}
                          value={opt}
                          onChange={(e) => handleOptionChange(idx, e.target.value)}
                          maxLength={100}
                          required
                        />
                        {options.length > 2 && (
                          <button
                            type="button"
                            className="builder-option-remove-btn"
                            onClick={() => handleRemoveOption(idx)}
                            title="Remove this option"
                            aria-label={`Remove option ${idx + 1}`}
                          >
                            <XIcon size={15} />
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>

                {options.length < 10 && (
                  <button
                    type="button"
                    className="builder-add-option-btn"
                    onClick={handleAddOption}
                  >
                    <PlusIcon size={16} />
                    <span>Add option</span>
                  </button>
                )}
              </div>

              {/* SECTION: SETTINGS */}
              <div className="form-section-block mt-28">
                <h3 className="builder-section-heading">Settings</h3>

                <div className="builder-settings-list">
                  <Toggle
                    id="one-vote-toggle"
                    checked={oneVotePerParticipant}
                    onChange={setOneVotePerParticipant}
                    label="Allow one vote per participant"
                    description="Prevents repeated votes from the same device session"
                  />

                  <Toggle
                    id="show-results-toggle"
                    checked={showResultsAfterVoting}
                    onChange={setShowResultsAfterVoting}
                    label="Show results after voting"
                    description="Participants can view the live tally immediately after voting"
                  />
                </div>
              </div>

              {/* SECTION: POLL STATUS */}
              <div className="form-section-block mt-24">
                <span className="builder-field-label">Poll status</span>
                <div className="status-selector-grid">
                  <label
                    className={`status-selector-card ${initialStatus === 'active' ? 'selected' : ''}`}
                  >
                    <input
                      type="radio"
                      name="initialStatus"
                      value="active"
                      checked={initialStatus === 'active'}
                      onChange={() => setInitialStatus('active')}
                      className="status-hidden-radio"
                    />
                    <div className="status-card-body">
                      <div className="status-radio-head">
                        <span className={`status-radio-indicator ${initialStatus === 'active' ? 'active' : ''}`} />
                        <span className="status-title-text">Active</span>
                      </div>
                      <span className="status-desc-text">Open for voting immediately</span>
                    </div>
                  </label>

                  <label
                    className={`status-selector-card ${initialStatus === 'closed' ? 'selected' : ''}`}
                  >
                    <input
                      type="radio"
                      name="initialStatus"
                      value="closed"
                      checked={initialStatus === 'closed'}
                      onChange={() => setInitialStatus('closed')}
                      className="status-hidden-radio"
                    />
                    <div className="status-card-body">
                      <div className="status-radio-head">
                        <span className={`status-radio-indicator ${initialStatus === 'closed' ? 'active' : ''}`} />
                        <span className="status-title-text">Closed</span>
                      </div>
                      <span className="status-desc-text">Draft mode (open manually later)</span>
                    </div>
                  </label>
                </div>
              </div>

              {/* 10. CREATE POLL ACTION BAR */}
              <div className="builder-action-bar-footer">
                <Button
                  type="button"
                  variant="ghost"
                  size="md"
                  onClick={() => navigate('/dashboard')}
                  disabled={loading}
                >
                  Cancel
                </Button>

                <div className="action-bar-right-group">
                  <Button
                    type="button"
                    variant="secondary"
                    size="md"
                    className="mobile-preview-btn"
                    onClick={scrollToPreview}
                    icon={<EyeIcon size={16} />}
                  >
                    Preview
                  </Button>

                  <Button
                    type="submit"
                    variant="primary"
                    size="lg"
                    isLoading={loading}
                    className="builder-submit-cta"
                  >
                    Create Poll
                  </Button>
                </div>
              </div>
            </form>
          </div>
        </div>

        {/* ==================================================================
            RIGHT COLUMN: LIVE VOTER PREVIEW (Synchronous Real-Time Preview)
            ================================================================== */}
        <div id="live-preview-section" className="builder-preview-column">
          <div className="sticky-preview-wrapper">
            <div className="preview-column-top-badge">
              <span className="preview-label-tag">LIVE VOTER PREVIEW</span>
              <span className="preview-live-indicator">
                <span className="live-dot-green" />
                Realtime Preview
              </span>
            </div>

            <div className="voter-card-preview-shell">
              {/* Preview Header */}
              <div className="voter-preview-header">
                <PollFlowLogo height={22} showText={true} />
                <span className="voter-preview-pill">
                  {initialStatus === 'active' ? '● LIVE' : 'DRAFT'}
                </span>
              </div>

              {/* Preview Question */}
              <h2 className="voter-preview-question">
                {question.trim() ? (
                  question.trim()
                ) : (
                  <span style={{ color: 'var(--text-light)', fontStyle: 'italic', fontWeight: 500 }}>
                    Your question will appear here...
                  </span>
                )}
              </h2>

              <p className="voter-preview-prompt">
                Select one option to cast your vote:
              </p>

              {/* Preview Options with clickable radio rings */}
              <div className="voter-preview-options-list">
                {options.map((opt, idx) => {
                  const hasText = opt.trim().length > 0
                  const isChecked = previewSelectedIdx === idx

                  return (
                    <div
                      key={idx}
                      className={`voter-preview-option-item ${isChecked ? 'selected' : ''}`}
                      onClick={() => setPreviewSelectedIdx(idx)}
                      role="button"
                      tabIndex={0}
                    >
                      <span className={`custom-radio-circle ${isChecked ? 'checked' : ''}`}>
                        {isChecked && <span className="custom-radio-inner" />}
                      </span>
                      <span
                        className="voter-preview-opt-text"
                        style={!hasText ? { color: 'var(--text-light)', fontStyle: 'italic' } : {}}
                      >
                        {hasText ? opt.trim() : `Option ${idx + 1}`}
                      </span>
                    </div>
                  )
                })}
              </div>

              {/* Preview Submit Button */}
              <div className="voter-preview-action-block">
                <button
                  type="button"
                  className="voter-preview-submit-btn"
                  onClick={() => alert('This is a live preview. Your actual voters will submit their vote here!')}
                >
                  Submit Vote
                </button>
              </div>

              <div className="voter-preview-footer">
                <span>Powered by PollFlow</span>
              </div>
            </div>

            <div className="preview-sync-note">
              <span>This preview updates synchronously as you type your question & options.</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
