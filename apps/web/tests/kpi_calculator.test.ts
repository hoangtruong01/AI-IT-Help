// QA Unit Test Suite for Frontend BI & SLA Math

import { describe, expect, it } from 'vitest'

export function calculateSLACompliance(totalTickets: number, resolvedWithinSLA: number): number {
  if (totalTickets <= 0) return 0.0
  const pct = (resolvedWithinSLA / totalTickets) * 100
  return Number.isFinite(pct) ? Math.round(pct * 100) / 100 : 0.0
}

export function calculateMTTRImprovement(previousMTTR: number, currentMTTR: number): number {
  if (previousMTTR <= 0) return 0.0
  const imp = ((previousMTTR - currentMTTR) / previousMTTR) * 100
  return Number.isFinite(imp) ? Math.round(imp * 10) / 10 : 0.0
}

export function maskCreditCard(cardNumber: string): string {
  const digits = cardNumber.replace(/\D/g, '')
  if (digits.length < 13 || digits.length > 16) return cardNumber
  return `****-****-****-${digits.slice(-4)}`
}

describe('KPI calculators', () => {
  it('returns a zero-safe SLA rate', () => {
    expect(calculateSLACompliance(0, 0)).toBe(0)
    expect(calculateSLACompliance(100, 97)).toBe(97)
  })

  it('calculates MTTR improvement', () => {
    expect(calculateMTTRImprovement(38.87, 31.8)).toBeGreaterThan(0)
  })

  it('masks valid card numbers', () => {
    expect(maskCreditCard('4111 2222 3333 4444')).toBe('****-****-****-4444')
  })
})
