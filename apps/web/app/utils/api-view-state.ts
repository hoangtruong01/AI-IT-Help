export type ApiViewState = 'loading' | 'ready' | 'empty' | 'forbidden' | 'unavailable'

interface ApiErrorShape {
  status?: number
  statusCode?: number
  response?: { status?: number }
  data?: { statusCode?: number, status?: number }
}

export function apiErrorStatus(error: unknown): number | undefined {
  const candidate = error as ApiErrorShape | null | undefined
  return candidate?.statusCode
    ?? candidate?.status
    ?? candidate?.response?.status
    ?? candidate?.data?.statusCode
    ?? candidate?.data?.status
}

export function classifyApiError(error: unknown): Extract<ApiViewState, 'forbidden' | 'unavailable'> {
  return apiErrorStatus(error) === 403 ? 'forbidden' : 'unavailable'
}

export function dataViewState(data: readonly unknown[] | null | undefined): Extract<ApiViewState, 'ready' | 'empty'> {
  return data && data.length > 0 ? 'ready' : 'empty'
}
