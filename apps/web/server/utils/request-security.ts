import type { H3Event } from 'h3'
import { createError, getRequestHeader, getRequestURL } from 'h3'
import { isAllowedBrowserOrigin } from '../../app/utils/request-origin'

/**
 * Protect cookie-authenticated BFF mutations from CSRF. SameSite cookies are a
 * useful browser control, but an exact Origin check also blocks same-site
 * sibling subdomains and makes the trust boundary explicit.
 */
export function assertSameOrigin(event: H3Event): void {
  const config = useRuntimeConfig(event)
  const configuredOrigins = String(config.allowedOrigins || '')
    .split(',')
    .map(value => value.trim())
    .filter(Boolean)

  const requestOrigin = getRequestURL(event).origin
  const origin = getRequestHeader(event, 'origin')
  const fetchSite = getRequestHeader(event, 'sec-fetch-site')

  if (!isAllowedBrowserOrigin(origin, requestOrigin, fetchSite, configuredOrigins)) {
    throw createError({
      statusCode: 403,
      statusMessage: 'Cross-origin authentication request rejected'
    })
  }
}
