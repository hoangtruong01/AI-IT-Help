function normalizeOrigin(value: string): string | null {
  try {
    const url = new URL(value)
    if (!['http:', 'https:'].includes(url.protocol)) return null
    return url.origin
  } catch {
    return null
  }
}

export function isAllowedBrowserOrigin(
  origin: string | undefined,
  requestOrigin: string,
  fetchSite: string | undefined,
  configuredOrigins: string[] = []
): boolean {
  if (fetchSite?.toLowerCase() === 'cross-site' || !origin) return false

  const normalizedOrigin = normalizeOrigin(origin)
  if (!normalizedOrigin) return false

  const allowed = new Set(
    [requestOrigin, ...configuredOrigins]
      .map(normalizeOrigin)
      .filter((value): value is string => value !== null)
  )
  return allowed.has(normalizedOrigin)
}
