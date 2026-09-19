import { apiClient } from './client'

export const pollsApi = {
  // Public: Get poll metadata and options for voting
  async getPoll(id) {
    const res = await apiClient(`/polls/${id}`)
    return res.data
  },

  // Public: Get current tally and results
  async getPollResults(id) {
    const res = await apiClient(`/polls/${id}/results`)
    return res.data
  },

  // Public: Cast a vote with X-Voter-ID header
  async vote(pollId, optionId, voterId) {
    const res = await apiClient(`/polls/${pollId}/vote`, {
      method: 'POST',
      headers: {
        'X-Voter-ID': voterId,
      },
      body: JSON.stringify({ optionId }),
    })
    return res.data
  },

  // Protected: Create a new poll
  async createPoll(question, options) {
    const res = await apiClient('/polls', {
      method: 'POST',
      body: JSON.stringify({ question, options }),
    })
    return res.data
  },

  // Protected: Get list of creator's polls
  async getMyPolls() {
    const res = await apiClient('/my/polls')
    return res.data || []
  },

  // Protected: Delete a poll
  async deletePoll(id) {
    const res = await apiClient(`/polls/${id}`, {
      method: 'DELETE',
    })
    return res.data
  },

  // Protected: Set poll status (Open / Closed)
  async setPollStatus(id, isActive) {
    const res = await apiClient(`/polls/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ isActive }),
    })
    return res.data
  },
}
