import { ref } from 'vue'
import { authApi } from '../api'

export function useCaptcha() {
  const captcha = ref(null)
  const loading = ref(false)
  const error = ref(null)

  async function fetchCaptcha() {
    try {
      loading.value = true
      error.value = null
      captcha.value = await authApi.getCaptcha()
      return captcha.value
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to fetch captcha'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    captcha,
    loading,
    error,
    fetchCaptcha
  }
}

export function useVerificationCode() {
  const loading = ref(false)
  const error = ref(null)
  const sent = ref(false)
  const expiresIn = ref(0)

  async function sendCode(email, captchaId, captchaCode) {
    try {
      loading.value = true
      error.value = null
      sent.value = false
      
      const response = await authApi.sendVerificationCode(email, captchaId, captchaCode)
      
      sent.value = true
      expiresIn.value = response.expires_in
      
      return response
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to send verification code'
      throw err
    } finally {
      loading.value = false
    }
  }

  function reset() {
    sent.value = false
    expiresIn.value = 0
    error.value = null
  }

  return {
    loading,
    error,
    sent,
    expiresIn,
    sendCode,
    reset
  }
}
