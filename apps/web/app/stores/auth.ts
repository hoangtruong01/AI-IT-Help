import { defineStore } from 'pinia'
import type { User } from '~/types'

interface BffAuthResponse {
  access_token: string
  user: User
}

export const useAuthStore = defineStore('auth', () => {
  // In-memory access token storage (Protected against XSS storage theft)
  const token = ref<string | null>(null)
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const isInitialized = ref(false)

  const isAuthenticated = computed(() => !!token.value)
  const role = computed(() => user.value?.role || 'ROLE_GUEST')
  const isAdmin = computed(() => user.value?.role === 'ROLE_ADMIN')

  async function login(email: string, password: string): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      // Call Nuxt BFF Server Route which sets HttpOnly refresh cookie
      const res = await $fetch<BffAuthResponse>('/api/auth/login', {
        method: 'POST',
        body: { email, password }
      })

      token.value = res.access_token
      user.value = res.user
      isInitialized.value = true
      return true
    } catch (err: unknown) {
      const errorObj = err as { data?: { message?: string, statusMessage?: string }, statusMessage?: string, message?: string }
      error.value = errorObj?.data?.statusMessage || errorObj?.data?.message || errorObj?.statusMessage || errorObj?.message || 'Login failed. Please check your credentials.'
      return false
    } finally {
      loading.value = false
    }
  }

  async function refreshSession(): Promise<string | null> {
    try {
      const res = await $fetch<BffAuthResponse>('/api/auth/refresh', {
        method: 'POST'
      })

      token.value = res.access_token
      user.value = res.user
      isInitialized.value = true
      return res.access_token
    } catch (err) {
      token.value = null
      user.value = null
      isInitialized.value = true
      throw err
    }
  }

  async function fetchCurrentUser(): Promise<void> {
    if (!token.value) return
    try {
      const res = await $fetch<User>('/api/auth/me', {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })
      user.value = res
    } catch {
      // If token expired or invalid, attempt refresh or logout
      try {
        await refreshSession()
      } catch {
        await logout()
      }
    }
  }

  async function logout() {
    try {
      await $fetch('/api/auth/logout', {
        method: 'POST'
      })
    } catch (e) {
      console.warn('Logout revocation error:', e)
    } finally {
      token.value = null
      user.value = null
      isInitialized.value = true
      navigateTo('/login')
    }
  }

  async function initSession(): Promise<void> {
    if (isInitialized.value) return
    try {
      await refreshSession()
    } catch {
      // No active session cookie; remain guest
      token.value = null
      user.value = null
      isInitialized.value = true
    }
  }

  return {
    user,
    token,
    loading,
    error,
    isInitialized,
    isAuthenticated,
    role,
    isAdmin,
    login,
    logout,
    refreshSession,
    fetchCurrentUser,
    initSession
  }
})
