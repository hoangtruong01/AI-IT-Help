import { defineEventHandler, getHeader, createError } from 'h3'
import { getBackendApiBase } from '../../utils/backend'

export default defineEventHandler(async (event) => {
  const authHeader = getHeader(event, 'authorization')
  if (!authHeader) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Missing authorization header'
    })
  }

  const apiBase = getBackendApiBase(event)

  try {
    const res = await $fetch(`${apiBase}/api/v1/auth/me`, {
      method: 'GET',
      headers: {
        Authorization: authHeader
      }
    })
    return res
  } catch (err: unknown) {
    const errorObj = err as { statusCode?: number, data?: { error?: { message?: string } }, message?: string }
    throw createError({
      statusCode: errorObj.statusCode || 401,
      statusMessage: errorObj.data?.error?.message || errorObj.message || 'Failed to fetch user profile'
    })
  }
})
