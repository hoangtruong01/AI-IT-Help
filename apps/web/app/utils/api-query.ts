export type ApiQueryValue = string | number | boolean | null | undefined
export type ApiQuery = Record<string, ApiQueryValue>

/** Build a deterministic, encoded query string for GET requests. */
export function withQuery(url: string, query?: ApiQuery): string {
  if (!query) return url

  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === '') continue
    if (typeof value === 'object') {
      throw new TypeError(`GET query parameter "${key}" must be a primitive value`)
    }
    search.append(key, String(value))
  }

  const encoded = search.toString()
  if (!encoded) return url
  return `${url}${url.includes('?') ? '&' : '?'}${encoded}`
}
