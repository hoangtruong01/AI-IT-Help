/**
 * API client composable.
 * Wraps $fetch with base URL and auth token handling.
 */
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiUrl

  async function request<T>(
    url: string,
    options: Parameters<typeof $fetch>[1] = {}
  ): Promise<T> {
    // TODO: Add auth token from auth store when implemented
    // const authStore = useAuthStore()
    // const headers = authStore.token
    //   ? { Authorization: `Bearer ${authStore.token}`, ...options.headers }
    //   : options.headers

    return $fetch<T>(url, {
      baseURL,
      ...options
    })
  }

  return {
    get: <T>(url: string, params?: Record<string, unknown>) =>
      request<T>(url, { method: 'GET', params }),

    post: <T>(url: string, body?: Record<string, unknown>) =>
      request<T>(url, { method: 'POST', body }),

    put: <T>(url: string, body?: Record<string, unknown>) =>
      request<T>(url, { method: 'PUT', body }),

    patch: <T>(url: string, body?: Record<string, unknown>) =>
      request<T>(url, { method: 'PATCH', body }),

    delete: <T>(url: string) =>
      request<T>(url, { method: 'DELETE' })
  }
}
