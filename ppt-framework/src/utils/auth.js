import { useAuthStore } from '../stores/auth'

export function requireAuth() {
  const authStore = useAuthStore()
  
  if (!authStore.isAuthenticated) {
    const returnUrl = encodeURIComponent(globalThis.location.pathname + globalThis.location.search)
    globalThis.location.href = `/login?return=${returnUrl}`
    return false
  }
  
  return true
}

export function redirectIfAuthenticated(redirectTo = '/') {
  const authStore = useAuthStore()
  
  if (authStore.isAuthenticated) {
    globalThis.location.href = redirectTo
    return true
  }
  
  return false
}

export function setupAuthInterceptor() {
  const authStore = useAuthStore()
  
  globalThis.addEventListener('storage', (event) => {
    if (event.key === 'accessToken' && !event.newValue) {
      authStore.clearAuth()
      globalThis.location.href = '/login'
    }
  })
}
