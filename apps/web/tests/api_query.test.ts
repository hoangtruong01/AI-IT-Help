import { describe, expect, it } from 'vitest'
import { withQuery } from '../app/utils/api-query'

describe('withQuery', () => {
  it('builds the exact encoded GET URL', () => {
    expect(withQuery('/api/v1/tickets', {
      page: 1,
      page_size: 100,
      search: 'printer & toner'
    })).toBe('/api/v1/tickets?page=1&page_size=100&search=printer+%26+toner')
  })

  it('omits empty optional values but keeps false and zero', () => {
    expect(withQuery('/reports?format=csv', {
      range: undefined,
      search: '',
      active: false,
      offset: 0
    })).toBe('/reports?format=csv&active=false&offset=0')
  })

  it('rejects the legacy nested params shape', () => {
    expect(() => withQuery('/tickets', {
      params: { page: 1 }
    } as never)).toThrow(/primitive value/)
  })
})
