import { test, expect } from '@playwright/test'
import { bearer, e2eConfig, loginInBrowser, loginViaGateway } from '../support/auth'

test('stale browser edit displays a real 409 conflict message', async ({ page, request }) => {
  const employee = await loginViaGateway(request, e2eConfig.employee)
  const agent = await loginViaGateway(request, e2eConfig.agent)
  const title = `Playwright conflict ${Date.now()}`

  const createdResponse = await request.post(`${e2eConfig.gatewayBaseURL}/api/v1/tickets`, {
    headers: bearer(employee.access_token),
    data: { title, description: 'Optimistic concurrency browser verification', category: 'Software', priority: 'HIGH' }
  })
  expect(createdResponse.status(), await createdResponse.text()).toBe(201)
  const created = await createdResponse.json()

  const assignedResponse = await request.patch(`${e2eConfig.gatewayBaseURL}/api/v1/tickets/${created.id}/assign`, {
    headers: bearer(agent.access_token),
    data: { assignee_id: agent.user.id, assignee_name: 'Playwright Agent', version: created.version }
  })
  expect(assignedResponse.status(), await assignedResponse.text()).toBe(200)
  const assigned = await assignedResponse.json()

  await loginInBrowser(page, e2eConfig.agent)
  await page.goto('/helpdesk')
  await page.getByRole('row').filter({ hasText: title }).getByRole('button', { name: /View/ }).click()
  await expect(page.getByTestId('ticket-detail-modal')).toBeVisible()

  const externalUpdate = await request.patch(`${e2eConfig.gatewayBaseURL}/api/v1/tickets/${created.id}/status`, {
    headers: bearer(agent.access_token),
    data: { status: 'IN_PROGRESS', notes: 'Concurrent writer', version: assigned.version }
  })
  expect(externalUpdate.status(), await externalUpdate.text()).toBe(200)

  await page.getByRole('button', { name: /Start Work/ }).click()
  await expect(page.getByTestId('ticket-action-error')).toContainText('updated by another user')
})
