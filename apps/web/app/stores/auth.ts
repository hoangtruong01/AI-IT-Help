import { defineStore } from 'pinia'
import type { User, AuthResponse } from '~/types'

export const useAuthStore = defineStore('auth', () => {
  const config = useRuntimeConfig()
  const apiBase = (config.public.apiUrl || 'http://localhost:8080').replace(/\/$/, '')

  const tokenCookie = useCookie<string | null>('eomp_token', {
    maxAge: 60 * 60, // access token lifetime
    sameSite: 'lax',
    path: '/',
    secure: import.meta.env.PROD
  })
  const refreshTokenCookie = useCookie<string | null>('eomp_refresh_token', {
    maxAge: 60 * 60 * 24 * 7,
    sameSite: 'lax',
    path: '/',
    secure: import.meta.env.PROD
  })

  const token = ref<string | null>(tokenCookie.value || null)
  const refreshToken = ref<string | null>(refreshTokenCookie.value || null)
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)
  const role = computed(() => user.value?.role || 'ROLE_GUEST')
  const isAdmin = computed(() => user.value?.role === 'ROLE_ADMIN')

  async function login(email: string, password: string): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      const res = await $fetch<AuthResponse>(`${apiBase}/api/v1/auth/login`, {
        method: 'POST',
        body: { email, password }
      })

      token.value = res.access_token
      refreshToken.value = res.refresh_token
      tokenCookie.value = res.access_token
      refreshTokenCookie.value = res.refresh_token
      user.value = res.user
      return true
    } catch (err: unknown) {
      const errorObj = err as { data?: { error?: { message?: string } }, message?: string }
      error.value = errorObj?.data?.error?.message || errorObj?.message || 'Login failed. Please check your credentials.'
      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchCurrentUser(): Promise<void> {
    if (!token.value) return
    try {
      const res = await $fetch<User>(`${apiBase}/api/v1/auth/me`, {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })
      user.value = res
    } catch {
      // If token expired or invalid, clear auth
      logout()
    }
  }

  async function logout() {
    const currentRefreshToken = refreshToken.value
    if (currentRefreshToken) {
      try {
        await $fetch(`${apiBase}/api/v1/auth/logout`, {
          method: 'POST',
          body: { refresh_token: currentRefreshToken }
        })
      } catch (e) {
        console.warn('Backend logout revocation error:', e)
      }
    }
    token.value = null
    refreshToken.value = null
    tokenCookie.value = null
    refreshTokenCookie.value = null
    user.value = null
    navigateTo('/login')
  }

  // Initialize user profile if token exists on load
  if (token.value && !user.value) {
    fetchCurrentUser()
  }

  return {
    user,
    token,
    refreshToken,
    loading,
    error,
    isAuthenticated,
    role,
    isAdmin,
    login,
    logout,
    fetchCurrentUser
  }
})
