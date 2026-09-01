import { useAuthStore } from '~/stores/auth'
import type { ApiQuery } from '~/utils/api-query'
import { withQuery } from '~/utils/api-query'
import { runWithRefreshMutex } from '~/utils/refresh-mutex'

/**
 * API client composable.
 * Wraps $fetch with base URL, JWT token injection, and 401 Refresh Mutex.
 */
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiUrl || 'http://localhost:8080'

  async function executeRequest<T>(
    url: string,
    options: Parameters<typeof $fetch>[1] = {},
    isRetry = false
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
      const errorObj = err as { statusCode?: number, status?: number, response?: { status?: number } }
      const statusCode = errorObj?.statusCode || errorObj?.status || errorObj?.response?.status

      // Handle 401 Unauthorized with Refresh Mutex
      if (statusCode === 401 && !isRetry && !url.includes('/api/auth/')) {
        try {
          const newToken = await runWithRefreshMutex(authStore, async () => {
            try {
              return await authStore.refreshSession()
            } catch (refreshError) {
              // This branch belongs to the shared promise, so concurrent 401s
              // clear the session and navigate only once.
              await authStore.logout()
              throw refreshError
            }
          })
          if (newToken) {
            const retryHeaders = {
              ...headers,
              Authorization: `Bearer ${newToken}`
            }
            return await executeRequest<T>(url, { ...options, headers: retryHeaders }, true)
          }
        } catch {
          throw err
        }
      }

      if (statusCode === 401 && isRetry) {
        await authStore.logout()
      }

      throw err
    }
  }

  return {
    get: <T>(url: string, params?: ApiQuery) =>
      executeRequest<T>(withQuery(url, params), { method: 'GET' }),

    post: <T>(url: string, body?: unknown) =>
      executeRequest<T>(url, { method: 'POST', body: body as Record<string, unknown> }),

    put: <T>(url: string, body?: unknown) =>
      executeRequest<T>(url, { method: 'PUT', body: body as Record<string, unknown> }),

    patch: <T>(url: string, body?: unknown) =>
      executeRequest<T>(url, { method: 'PATCH', body: body as Record<string, unknown> }),

    delete: <T>(url: string) =>
      executeRequest<T>(url, { method: 'DELETE' })
  }
}
