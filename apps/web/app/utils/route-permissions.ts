/**
 * Route RBAC Configuration mapping route prefixes to allowed roles.
 */
export const ROUTE_PERMISSIONS: Record<string, string[]> = {
  '/audit': ['ROLE_ADMIN', 'ROLE_MANAGER'],
  '/reports': ['ROLE_ADMIN', 'ROLE_MANAGER'],
  '/monitoring': ['ROLE_ADMIN', 'ROLE_MANAGER'],
  '/changes': ['ROLE_ADMIN', 'ROLE_MANAGER'],
  '/problems': ['ROLE_ADMIN', 'ROLE_MANAGER', 'ROLE_AGENT'],
  '/assets': ['ROLE_ADMIN', 'ROLE_MANAGER', 'ROLE_AGENT']
}

/**
 * Evaluates whether a given user role has permission to access a target path.
 */
export function isRouteAllowedForRole(path: string, role: string): boolean {
  for (const [prefix, allowedRoles] of Object.entries(ROUTE_PERMISSIONS)) {
    if (path === prefix || path.startsWith(prefix + '/')) {
      return allowedRoles.includes(role)
    }
  }
  return true
}

export function safeLocalRedirect(value: unknown, fallback = '/'): string {
  if (typeof value !== 'string') return fallback
  if (!value.startsWith('/') || value.startsWith('//') || value.includes('\\')) return fallback
  return value
}
