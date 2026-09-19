import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { pollsApi } from '../api/polls'
import { PollVoteCard } from '../components/poll/PollVoteCard'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { Alert } from '../components/common/Alert'
import { Button } from '../components/common/Button'
import { ZapIcon, RefreshIcon } from '../components/icons/Icons'

function getOrCreateVoterId() {
  let voterId = localStorage.getItem('livepoll_voter_id')
  if (!voterId) {
    voterId = 'voter_' + Math.random().toString(36).substring(2, 11) + '_' + Date.now().toString(36)
    localStorage.setItem('livepoll_voter_id', voterId)
  }
  return voterId
}

export function PollVotePage() {
  const { id, pollId } = useParams()
  const effectiveId = id || pollId

  const [poll, setPoll] = useState(null)
  const [loading, setLoading] = useState(true)
  const [voting, setVoting] = useState(false)
  const [error, setError] = useState(null)
  const [voteError, setVoteError] = useState(null)
  const [hasVoted, setHasVoted] = useState(false)
  const [votedOptionId, setVotedOptionId] = useState(null)

  // Check if voter already casted a vote for this poll locally
  useEffect(() => {
    if (!effectiveId) return
    const votedMap = JSON.parse(localStorage.getItem('livepoll_voted_polls') || '{}')
    if (votedMap[effectiveId]) {
      setHasVoted(true)
      setVotedOptionId(votedMap[effectiveId])
    }
  }, [effectiveId])

  const fetchPoll = useCallback(async () => {
    if (!effectiveId) return
    setLoading(true)
    setError(null)
    try {
      const data = await pollsApi.getPoll(effectiveId)
      setPoll(data)
    } catch (err) {
      setError(err.message || 'Poll could not be found or has expired.')
    } finally {
      setLoading(false)
    }
  }, [effectiveId])

  useEffect(() => {
    fetchPoll()
  }, [fetchPoll])

  const handleVote = async (optionId) => {
    if (!poll || voting || hasVoted || !effectiveId) return

    setVoting(true)
    setVoteError(null)

    const voterId = getOrCreateVoterId()

    try {
      await pollsApi.vote(effectiveId, optionId, voterId)
      
      // Save local vote record
      const votedMap = JSON.parse(localStorage.getItem('livepoll_voted_polls') || '{}')
      votedMap[effectiveId] = optionId
      localStorage.setItem('livepoll_voted_polls', JSON.stringify(votedMap))

      setHasVoted(true)
      setVotedOptionId(optionId)
    } catch (err) {
      if (err.message?.toLowerCase().includes('already voted')) {
        setHasVoted(true)
        setVoteError('You have already submitted a vote for this poll.')
      } else {
        setVoteError(err.message || 'Failed to submit vote.')
      }
    } finally {
      setVoting(false)
    }
  }

  // Session reset for demo convenience
  const handleResetSession = () => {
    const newVoter = 'voter_' + Math.random().toString(36).substring(2, 11) + '_' + Date.now().toString(36)
    localStorage.setItem('livepoll_voter_id', newVoter)
    if (effectiveId) {
      const votedMap = JSON.parse(localStorage.getItem('livepoll_voted_polls') || '{}')
      delete votedMap[effectiveId]
      localStorage.setItem('livepoll_voted_polls', JSON.stringify(votedMap))
    }
    setHasVoted(false)
    setVotedOptionId(null)
    setVoteError(null)
  }

  if (loading) {
    return (
      <div className="page-center-container">
        <LoadingSpinner text="Loading poll question..." size="lg" />
      </div>
    )
  }

  if (error || !poll) {
    return (
      <div className="page-center-container">
        <div className="error-panel glass-panel text-center max-w-md">
          <h2>Poll Unavailable</h2>
          <p className="text-muted mt-8 mb-20">{error || 'This poll does not exist or has been removed.'}</p>
          <div className="flex gap-12 justify-center">
            <Button variant="outline" size="sm" onClick={fetchPoll} icon={<RefreshIcon size={16} />}>
              Retry
            </Button>
            <Link to="/">
              <Button variant="primary" size="sm">
                Go to Home
              </Button>
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="public-vote-page-container">
      <div className="public-vote-shell">
        <PollVoteCard
          poll={poll}
          onVote={handleVote}
          isVoting={voting}
          hasVoted={hasVoted}
          votedOptionId={votedOptionId}
          error={voteError}
        />
      </div>
    </div>
  )
}
