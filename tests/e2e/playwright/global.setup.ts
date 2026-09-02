import { request, type FullConfig } from '@playwright/test'

const requiredCredentials = [
  'E2E_ADMIN_EMAIL', 'E2E_ADMIN_PASSWORD',
  'E2E_EMPLOYEE_EMAIL', 'E2E_EMPLOYEE_PASSWORD',
  'E2E_AGENT_EMAIL', 'E2E_AGENT_PASSWORD'
]

export default async function globalSetup(config: FullConfig) {
  const missing = requiredCredentials.filter(name => !process.env[name]?.trim())
  if (missing.length > 0) {
    throw new Error(`Playwright requires dedicated test credentials: ${missing.join(', ')}`)
  }

  const webBaseURL = String(config.projects[0]?.use?.baseURL || '')
  const gatewayBaseURL = process.env.E2E_GATEWAY_BASE_URL || 'http://127.0.0.1:8080'
  const client = await request.newContext()
  try {
    const [web, gateway] = await Promise.all([
      client.get(webBaseURL),
      client.get(`${gatewayBaseURL}/health`)
    ])
    if (!web.ok() || !gateway.ok()) {
      throw new Error(`E2E targets are not ready (web=${web.status()}, gateway=${gateway.status()})`)
    }
  } finally {
    await client.dispose()
  }
}
