import type { User } from '~/types'

/**
 * Auth store skeleton.
 * Will be fully implemented when auth service is ready.
 */
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(newUser: typeof user.value, newToken: string) {
    user.value = newUser
    token.value = newToken
  }

  function clearAuth() {
    user.value = null
    token.value = null
  }

  return {
    user,
    token,
    isAuthenticated,
    setAuth,
    clearAuth
  }
})
