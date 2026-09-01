import { defineEventHandler, getCookie, setCookie, deleteCookie, createError } from 'h3'
import { getBackendApiBase } from '../../utils/backend'
import { assertSameOrigin } from '../../utils/request-security'

export default defineEventHandler(async (event) => {
  assertSameOrigin(event)
  const refreshToken = getCookie(event, 'eomp_refresh_token')
  if (!refreshToken) {
    throw createError({
      statusCode: 401,
      statusMessage: 'No refresh token session cookie found'
    })
  }

  const apiBase = getBackendApiBase(event)

  try {
    const res = await $fetch<{
      access_token: string
      refresh_token: string
      user: Record<string, unknown>
    }>(`${apiBase}/api/v1/auth/refresh`, {
      method: 'POST',
      body: {
        refresh_token: refreshToken
      }
    })

    // Update rotated refresh token in HttpOnly cookie
    setCookie(event, 'eomp_refresh_token', res.refresh_token, {
      httpOnly: true,
      secure: !import.meta.dev,
      sameSite: 'strict',
      path: '/api/auth',
      maxAge: 60 * 60 * 24 * 7
    })

    return {
      access_token: res.access_token,
      user: res.user
    }
  } catch (err: unknown) {
    // If refresh token is invalid, expired, or replayed, purge cookie
    deleteCookie(event, 'eomp_refresh_token', {
      path: '/api/auth'
    })
    const errorObj = err as { statusCode?: number, data?: { error?: { message?: string } }, message?: string }
    throw createError({
      statusCode: errorObj.statusCode || 401,
      statusMessage: errorObj.data?.error?.message || errorObj.message || 'Session refresh failed'
    })
  }
})
