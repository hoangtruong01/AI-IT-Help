// QA Unit Test Suite for Frontend BI & SLA Math

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

// Test Runner Verification
function runTests() {
  // Test Case 9.2: Zero-safe compliance rate
  const zeroSLA = calculateSLACompliance(0, 0)
  if (zeroSLA !== 0.0) throw new Error(`Expected 0.0 for zero SLA, got ${zeroSLA}`)

  const normalSLA = calculateSLACompliance(100, 97)
  if (normalSLA !== 97.0) throw new Error(`Expected 97.0%, got ${normalSLA}`)

  // MTTR Improvement calculation
  const mttrImp = calculateMTTRImprovement(38.87, 31.8)
  if (mttrImp <= 0) throw new Error(`Expected positive MTTR improvement, got ${mttrImp}`)

  // Data Masking
  const masked = maskCreditCard('4111 2222 3333 4444')
  if (masked !== '****-****-****-4444') throw new Error(`Expected masked CC, got ${masked}`)

  return true
}

runTests()
