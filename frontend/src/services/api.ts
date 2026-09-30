import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import type { ApiResponse, HealthCheckData } from '@/types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

// Request Interceptor: Attach headers, Request ID, Auth token if available
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Generate simple client trace ID if not present
    if (!config.headers['X-Request-ID']) {
      config.headers['X-Request-ID'] = `client-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`
    }

    const token = localStorage.getItem('simak_access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Response Interceptor: Uniform error extraction & 401 handler
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error: AxiosError<ApiResponse>) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to /login if unauthorized (except on /auth/login)
      if (!error.config?.url?.includes('/auth/login')) {
        localStorage.removeItem('simak_access_token')
        localStorage.removeItem('simak_refresh_token')
        localStorage.removeItem('simak_user_cache')
        if (window.location.pathname !== '/login') {
          window.location.href = '/login'
        }
      }
    }

    const message = error.response?.data?.message || error.message || 'An unexpected error occurred'
    console.error(`[API Error] ${error.config?.url}:`, message)
    return Promise.reject(error)
  }
)

// Health Check API
export const healthService = {
  async checkHealth(): Promise<ApiResponse<HealthCheckData>> {
    const response = await apiClient.get<ApiResponse<HealthCheckData>>('/health')
    return response.data
  },
}
