import type { ApiResponse, ApiError } from '~/types/api'

export function useApi() {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()

  const baseURL = config.public.apiBase

  async function request<T>(
    path: string,
    options: RequestInit & { _retry?: boolean } = {},
  ): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (authStore.accessToken) {
      headers['Authorization'] = `Bearer ${authStore.accessToken}`
    }

    const res = await fetch(`${baseURL}${path}`, { ...options, headers, credentials: 'include' })

    if (res.status === 401 && !options._retry) {
      const refreshed = await authStore.refresh()
      if (refreshed) {
        return request<T>(path, { ...options, _retry: true })
      }
      await authStore.logout()
      await navigateTo('/login')
      throw new Error('Session expired')
    }

    const body = await res.json()

    if (!res.ok) {
      throw body as ApiError
    }

    return body as ApiResponse<T>
  }

  function get<T>(path: string) {
    return request<T>(path, { method: 'GET' })
  }

  function post<T>(path: string, body?: unknown) {
    return request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined })
  }

  function patch<T>(path: string, body?: unknown) {
    return request<T>(path, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined })
  }

  function del<T>(path: string) {
    return request<T>(path, { method: 'DELETE' })
  }

  async function upload<T>(path: string, formData: FormData): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {}
    if (authStore.accessToken) {
      headers['Authorization'] = `Bearer ${authStore.accessToken}`
    }
    const res = await fetch(`${baseURL}${path}`, {
      method: 'POST',
      body: formData,
      headers,
      credentials: 'include',
    })
    const body = await res.json()
    if (!res.ok) throw body as ApiError
    return body as ApiResponse<T>
  }

  return { get, post, patch, delete: del, upload, request }
}
