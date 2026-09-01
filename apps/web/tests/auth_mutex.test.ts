import { describe, it, expect } from 'vitest'
import { runWithRefreshMutex } from '../app/utils/refresh-mutex'

describe('runWithRefreshMutex', () => {
  it('deduplicates 5 concurrent refreshes for the same auth store', async () => {
    const store = {}
    let calls = 0
    const refresh = async () => {
      calls++
      await new Promise(resolve => setTimeout(resolve, 20))
      return 'new-access-token'
    }

    const results = await Promise.all(
      Array.from({ length: 5 }, () => runWithRefreshMutex(store, refresh))
    )

    expect(calls).toBe(1)
    expect(results).toEqual(Array(5).fill('new-access-token'))
  })

  it('does not share refresh results between SSR auth-store instances', async () => {
    const firstStore = {}
    const secondStore = {}

    const [first, second] = await Promise.all([
      runWithRefreshMutex(firstStore, async () => 'first-user-token'),
      runWithRefreshMutex(secondStore, async () => 'second-user-token')
    ])

    expect(first).toBe('first-user-token')
    expect(second).toBe('second-user-token')
  })

  it('clears a failed refresh so a later attempt can retry', async () => {
    const store = {}
    let calls = 0
    const refresh = async () => {
      calls++
      if (calls === 1) throw new Error('expired')
      return 'recovered-token'
    }

    await expect(runWithRefreshMutex(store, refresh)).rejects.toThrow('expired')
    await expect(runWithRefreshMutex(store, refresh)).resolves.toBe('recovered-token')
    expect(calls).toBe(2)
  })
})
