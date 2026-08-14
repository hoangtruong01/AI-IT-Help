/**
 * Common API response types.
 */

/** Standard API response wrapper */
export interface ApiResponse<T> {
  data: T
  message?: string
}

/** Paginated response */
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

/** Health check response */
export interface HealthResponse {
  status: 'ok' | 'error'
  service: string
  version: string
}

/** User session */
export interface User {
  id: string
  email: string
  name: string
  role: string
  avatar?: string
}

/** Navigation menu item */
export interface MenuItem {
  label: string
  icon?: string
  to?: string
  children?: MenuItem[]
  badge?: string | number
}
