import { test, expect } from '@playwright/test'
import { e2eConfig, loginInBrowser } from '../support/auth'

test('admin login renders role-specific navigation', async ({ page }) => {
  await loginInBrowser(page, e2eConfig.admin)
  await expect(page.getByText('ROLE_ADMIN', { exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Audit & Compliance' })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Reports & Analytics' })).toBeVisible()
})

test('employee route guard rejects privileged route', async ({ page }) => {
  await loginInBrowser(page, e2eConfig.employee)
  await expect(page.getByRole('link', { name: 'Audit & Compliance' })).toHaveCount(0)
  await page.goto('/audit')
  await expect(page).toHaveURL(/\/$/)
  await expect(page.getByText('ROLE_EMPLOYEE', { exact: true })).toBeVisible()
})
