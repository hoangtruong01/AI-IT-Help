import { describe, expect, it } from 'vitest'
import { apiErrorStatus, classifyApiError, dataViewState } from '../app/utils/api-view-state'

describe('API view state contract', () => {
  it.each([
    [{ statusCode: 403 }],
    [{ status: 403 }],
    [{ response: { status: 403 } }],
    [{ data: { statusCode: 403 } }]
  ])('classifies HTTP 403 as forbidden without treating it as empty', (error) => {
    expect(apiErrorStatus(error)).toBe(403)
    expect(classifyApiError(error)).toBe('forbidden')
  })

  it.each([
    [{ statusCode: 500 }],
    [{ response: { status: 503 } }],
    [new TypeError('fetch failed')]
  ])('classifies backend/network failures as unavailable', (error) => {
    expect(classifyApiError(error)).toBe('unavailable')
  })

  it('distinguishes a successful empty response from populated data', () => {
    expect(dataViewState([])).toBe('empty')
    expect(dataViewState(null)).toBe('empty')
    expect(dataViewState([{ id: 'record-1' }])).toBe('ready')
  })
})
