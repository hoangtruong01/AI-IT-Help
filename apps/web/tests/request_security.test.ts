import { describe, expect, it } from 'vitest'
import { isAllowedBrowserOrigin } from '../app/utils/request-origin'

describe('BFF request origin policy', () => {
  it('allows the exact application origin', () => {
    expect(isAllowedBrowserOrigin(
      'https://eomp.example.com',
      'https://eomp.example.com',
      'same-origin'
    )).toBe(true)
  })

  it('rejects cross-site and sibling-subdomain requests', () => {
    expect(isAllowedBrowserOrigin(
      'https://evil.example.net',
      'https://eomp.example.com',
      'cross-site'
    )).toBe(false)
    expect(isAllowedBrowserOrigin(
      'https://attacker.example.com',
      'https://eomp.example.com',
      'same-site'
    )).toBe(false)
  })

  it('requires Origin even when Fetch Metadata is absent', () => {
    expect(isAllowedBrowserOrigin(
      undefined,
      'https://eomp.example.com',
      undefined
    )).toBe(false)
  })

  it('accepts an explicitly configured alternate origin', () => {
    expect(isAllowedBrowserOrigin(
      'https://pilot.example.com',
      'https://eomp.example.com',
      'same-site',
      ['https://pilot.example.com']
    )).toBe(true)
  })
})
