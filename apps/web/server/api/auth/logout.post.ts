import { defineEventHandler, getCookie, deleteCookie } from 'h3'
import { getBackendApiBase } from '../../utils/backend'
import { assertSameOrigin } from '../../utils/request-security'

export default defineEventHandler(async (event) => {
  assertSameOrigin(event)
  const refreshToken = getCookie(event, 'eomp_refresh_token')
  const apiBase = getBackendApiBase(event)

  if (refreshToken) {
    try {
      await $fetch(`${apiBase}/api/v1/auth/logout`, {
        method: 'POST',
        body: { refresh_token: refreshToken }
      })
    } catch (e) {
      console.warn('Backend logout token revocation error:', e)
    }
  }

  // Clear HttpOnly refresh cookie
  deleteCookie(event, 'eomp_refresh_token', {
    path: '/api/auth'
  })

  return {
    success: true,
    message: 'Logged out successfully'
  }
})
