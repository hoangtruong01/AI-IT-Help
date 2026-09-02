import { expect, type APIRequestContext, type Page } from '@playwright/test'

export type Credentials = { email: string, password: string }

function required(name: string): string {
  const value = process.env[name]?.trim()
  if (!value) throw new Error(`${name} is required`)
  return value
}

export const e2eConfig = {
  webBaseURL: process.env.E2E_WEB_BASE_URL || 'http://127.0.0.1:3000',
  gatewayBaseURL: process.env.E2E_GATEWAY_BASE_URL || 'http://127.0.0.1:8080',
  admin: { email: required('E2E_ADMIN_EMAIL'), password: required('E2E_ADMIN_PASSWORD') },
  employee: { email: required('E2E_EMPLOYEE_EMAIL'), password: required('E2E_EMPLOYEE_PASSWORD') },
  agent: { email: required('E2E_AGENT_EMAIL'), password: required('E2E_AGENT_PASSWORD') }
}

export async function loginInBrowser(page: Page, credentials: Credentials) {
  await page.goto('/login')
  await page.locator('input[type="email"]').fill(credentials.email)
  await page.locator('input[type="password"]').fill(credentials.password)
  await page.getByRole('button', { name: 'Sign In to Account' }).click()
  await expect(page).not.toHaveURL(/\/login/)
}

export async function loginViaGateway(request: APIRequestContext, credentials: Credentials) {
  const response = await request.post(`${e2eConfig.gatewayBaseURL}/api/v1/auth/login`, {
    data: credentials
  })
  expect(response.status(), await response.text()).toBe(200)
  return response.json() as Promise<{
    access_token: string
    refresh_token: string
    user: { id: string, email: string, role: string, department_id?: string }
  }>
}

export function bearer(token: string) {
  return { Authorization: `Bearer ${token}` }
}
