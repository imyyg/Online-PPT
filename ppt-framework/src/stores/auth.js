import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const accessToken = ref(null)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!user.value && !!accessToken.value)

  function initFromStorage() {
    const storedUser = localStorage.getItem('user')
    const storedToken = localStorage.getItem('accessToken')
    
    if (storedUser && storedToken) {
      try {
        user.value = JSON.parse(storedUser)
        accessToken.value = storedToken
      } catch (e) {
        console.error('Failed to parse stored user data:', e)
        clearAuth()
      }
    }
  }

  function saveToStorage(userData, token) {
    user.value = userData
    accessToken.value = token
    localStorage.setItem('user', JSON.stringify(userData))
    localStorage.setItem('accessToken', token)
  }

  function clearAuth() {
    user.value = null
    accessToken.value = null
    localStorage.removeItem('user')
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  async function login(email, password) {
    try {
      loading.value = true
      error.value = null
      
      const response = await authApi.login(email, password)
      
      saveToStorage(response.user, response.accessToken)
      
      return response
    } catch (err) {
      error.value = err.response?.data?.message || 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(email, password, emailCode) {
    try {
      loading.value = true
      error.value = null
      
      const response = await authApi.register(email, password, emailCode)
      
      saveToStorage(response.user, response.accessToken)
      
      return response
    } catch (err) {
      error.value = err.response?.data?.message || 'Registration failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      loading.value = true
      error.value = null
      
      await authApi.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      clearAuth()
      loading.value = false
    }
  }

  async function refreshToken() {
    try {
      const response = await authApi.refresh()
      saveToStorage(response.user, response.accessToken)
      return response
    } catch (err) {
      clearAuth()
      throw err
    }
  }

  initFromStorage()

  return {
    user,
    accessToken,
    loading,
    error,
    isAuthenticated,
    login,
    register,
    logout,
    refreshToken,
    clearAuth
  }
})
