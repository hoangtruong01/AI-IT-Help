import { test, expect } from '@playwright/test'
import { e2eConfig, loginInBrowser } from '../support/auth'

test('concurrent unauthorized API responses trigger one refresh rotation', async ({ page }) => {
  let refreshRequests = 0
  const forced = new Set<string>()
  page.on('request', (request) => {
    if (request.url().includes('/api/auth/refresh')) refreshRequests++
  })
  await page.route(/\/api\/v1\/notifications(?:\/stats)?(?:\?.*)?$/, async (route) => {
    const key = new URL(route.request().url()).pathname
    if (!forced.has(key)) {
      forced.add(key)
      await route.fulfill({ status: 401, contentType: 'application/json', body: '{"error":{"message":"forced expiry"}}' })
      return
    }
    await route.continue()
  })

  await loginInBrowser(page, e2eConfig.employee)
  await expect.poll(() => forced.size).toBe(2)
  await expect.poll(() => refreshRequests).toBe(1)
})

test('logout clears the browser session and rejects refresh', async ({ page }) => {
  await loginInBrowser(page, e2eConfig.employee)
  await page.getByTitle('Sign Out').click()
  await expect(page).toHaveURL(/\/login/)

  const refresh = await page.request.post(`${e2eConfig.webBaseURL}/api/auth/refresh`, {
    headers: { Origin: e2eConfig.webBaseURL }
  })
  expect(refresh.status()).toBe(401)
})
