/**
 * Common API response and entity types for EOMP.
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
  page_size: number
  total_pages: number
}

/** Health check response */
export interface HealthResponse {
  status: 'ok' | 'error'
  service: string
  version: string
}

/** User entity */
export interface User {
  id: string
  email: string
  full_name: string
  role: 'ROLE_ADMIN' | 'ROLE_MANAGER' | 'ROLE_AGENT' | 'ROLE_EMPLOYEE' | string
  department_id?: string | null
  is_active: boolean
  created_at: string
}

/** Auth login/refresh response */
export interface AuthResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

/** Department entity */
export interface Department {
  id: string
  name: string
  code: string
  manager_id?: string | null
  parent_id?: string | null
  created_at: string
  updated_at: string
}

/** Employee entity */
export interface Employee {
  id: string
  user_id?: string | null
  first_name: string
  last_name: string
  full_name: string
  email: string
  phone?: string | null
  job_title: string
  department_id?: string | null
  department_name?: string | null
  department_code?: string | null
  manager_id?: string | null
  manager_name?: string | null
  status: 'ACTIVE' | 'ON_LEAVE' | 'PROBATION' | 'TERMINATED' | string
  location: string
  joined_at: string
  created_at: string
  updated_at: string
}

/** Create Employee Payload */
export interface CreateEmployeePayload {
  first_name: string
  last_name: string
  email: string
  phone?: string
  job_title: string
  department_id?: string
  manager_id?: string
  status?: string
  location?: string
  joined_at?: string
}

/** Navigation menu item */
export interface MenuItem {
  label: string
  icon?: string
  to?: string
  children?: MenuItem[]
  badge?: string | number
}
