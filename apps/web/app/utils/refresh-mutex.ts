const activeRefreshes = new WeakMap<object, Promise<string | null>>()

/**
 * Deduplicate refreshes per Pinia auth-store instance. Keying by store avoids
 * sharing one user's token across concurrent SSR request contexts.
 */
export function runWithRefreshMutex(
  key: object,
  refresh: () => Promise<string | null>
): Promise<string | null> {
  const active = activeRefreshes.get(key)
  if (active) return active

  const pending = Promise.resolve()
    .then(refresh)
    .finally(() => activeRefreshes.delete(key))

  activeRefreshes.set(key, pending)
  return pending
}
