import { defineEventHandler, proxyRequest } from 'h3'
import { getBackendApiBase } from '../../utils/backend'

// Same-origin proxy for browser API traffic. This keeps the browser CSP at
// connect-src 'self' while the Nitro server reaches Gateway on the private
// container network. Gateway remains the authentication/authorization boundary.
export default defineEventHandler(async (event) => {
  const apiBase = getBackendApiBase(event)
  return proxyRequest(event, `${apiBase}${event.path}`)
})
