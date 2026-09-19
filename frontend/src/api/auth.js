import { apiClient } from './client'

export const authApi = {
  async signup(email, password) {
    const res = await apiClient('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    return res.data
  },

  async login(email, password) {
    const res = await apiClient('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    return res.data
  },

  async getMe() {
    const res = await apiClient('/auth/me', {
      method: 'GET',
    })
    return res.data
  },
}
