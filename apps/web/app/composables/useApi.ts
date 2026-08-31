import { useAuthStore } from '~/stores/auth'
import type { ApiQuery } from '~/utils/api-query'
import { withQuery } from '~/utils/api-query'

/**
 * API client composable.
 * Wraps $fetch with base URL and automatic JWT auth token injection.
 */
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiUrl || 'http://localhost:8080'

  async function request<T>(
    url: string,
    options: Parameters<typeof $fetch>[1] = {}
  ): Promise<T> {
    const authStore = useAuthStore()

    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string> || {})
    }

    if (authStore.token) {
      headers.Authorization = `Bearer ${authStore.token}`
    }

    try {
      return await $fetch<T>(url, {
        baseURL,
        ...options,
        headers
      })
    } catch (err: unknown) {
      const errorObj = err as { statusCode?: number }
      if (errorObj?.statusCode === 401) {
        authStore.logout()
      }
      throw err
    }
  }

  return {
    get: <T>(url: string, params?: ApiQuery) =>
      request<T>(withQuery(url, params), { method: 'GET' }),

    post: <T>(url: string, body?: unknown) =>
      request<T>(url, { method: 'POST', body: body as Record<string, unknown> }),

    put: <T>(url: string, body?: unknown) =>
      request<T>(url, { method: 'PUT', body: body as Record<string, unknown> }),

    patch: <T>(url: string, body?: unknown) =>
      request<T>(url, { method: 'PATCH', body: body as Record<string, unknown> }),

    delete: <T>(url: string) =>
      request<T>(url, { method: 'DELETE' })
  }
}
