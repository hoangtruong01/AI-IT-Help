import { test, expect } from '@playwright/test'
import { e2eConfig, loginInBrowser } from '../support/auth'

test('employee creates a ticket and sees its number and SLA', async ({ page }) => {
  const title = `Playwright incident ${Date.now()}`
  await loginInBrowser(page, e2eConfig.employee)
  await page.goto('/helpdesk')
  await page.getByRole('button', { name: '+ Create New Ticket' }).click()
  const modal = page.getByText('Raise New Incident / Service Request').locator('..').locator('..')
  await modal.getByPlaceholder('Brief summary of the issue or request...').fill(title)
  await modal.getByPlaceholder('Provide exact symptoms, error codes, steps to reproduce, or requirements...').fill('Created by the real Chromium browser journey.')
  await modal.locator('select').nth(1).selectOption('HIGH')
  await modal.getByRole('button', { name: 'Submit Ticket' }).click()

  const row = page.getByRole('row').filter({ hasText: title })
  await expect(row).toBeVisible()
  await expect(row).toContainText(/TK-/)
  await expect(row).toContainText(/left|Breached|Resolved/i)
})
