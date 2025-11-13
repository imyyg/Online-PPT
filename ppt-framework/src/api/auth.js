import apiClient from './client'

export const authApi = {
  async getCaptcha() {
    const response = await apiClient.get('/auth/captcha')
    return response.data
  },

  async sendVerificationCode(email, captchaId, captchaCode) {
    const response = await apiClient.post('/auth/send-verification-code', {
      email,
      captcha_id: captchaId,
      captcha_code: captchaCode
    })
    return response.data
  },

  async register(email, password, emailCode) {
    const response = await apiClient.post('/auth/register', {
      email,
      password,
      email_code: emailCode
    })
    return response.data
  },

  async login(email, password) {
    const response = await apiClient.post('/auth/login', {
      email,
      password
    })
    return response.data
  },

  async logout() {
    const response = await apiClient.post('/auth/logout')
    return response.data
  },

  async refresh() {
    const refreshToken = localStorage.getItem('refreshToken')
    const response = await apiClient.post('/auth/refresh', {
      refreshToken
    })
    return response.data
  }
}
