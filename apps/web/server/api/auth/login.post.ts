import { defineEventHandler, readBody, setCookie, createError } from 'h3'
import { getBackendApiBase } from '../../utils/backend'
import { assertSameOrigin } from '../../utils/request-security'

export default defineEventHandler(async (event) => {
  assertSameOrigin(event)
  const body = await readBody(event)
  const apiBase = getBackendApiBase(event)

  try {
    const res = await $fetch<{
      access_token: string
      refresh_token: string
      user: Record<string, unknown>
    }>(`${apiBase}/api/v1/auth/login`, {
      method: 'POST',
      body: {
        email: body?.email,
        password: body?.password
      }
    })

    // Set HttpOnly, Secure, SameSite refresh token cookie on /api/auth path
    setCookie(event, 'eomp_refresh_token', res.refresh_token, {
      httpOnly: true,
      secure: !import.meta.dev,
      sameSite: 'strict',
      path: '/api/auth',
      maxAge: 60 * 60 * 24 * 7 // 7 days
    })

    // Return access token and user info in memory to frontend (refresh_token is kept in HttpOnly cookie)
    return {
      access_token: res.access_token,
      user: res.user
    }
  } catch (err: unknown) {
    const errorObj = err as { statusCode?: number, data?: { error?: { message?: string } }, message?: string }
    throw createError({
      statusCode: errorObj.statusCode || 401,
      statusMessage: errorObj.data?.error?.message || errorObj.message || 'Authentication failed'
    })
  }
})
