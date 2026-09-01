import type { H3Event } from 'h3'

export function getBackendApiBase(event: H3Event): string {
  const config = useRuntimeConfig(event)
  const configured = String(config.apiBaseUrl || '').trim()

  let url: URL
  try {
    url = new URL(configured)
  } catch {
    throw createError({
      statusCode: 503,
      statusMessage: 'Authentication service is unavailable'
    })
  }

  if (!['http:', 'https:'].includes(url.protocol)) {
    throw createError({
      statusCode: 503,
      statusMessage: 'Authentication service is unavailable'
    })
  }

  return url.toString().replace(/\/$/, '')
}
