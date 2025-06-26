import { useNotificationStore } from '../stores/notifications'

export interface ApiResponse<T = any> {
  data: T
  status: number
  message: string
}

export interface ApiError {
  message: string
  status?: number
  code?: string
}

export interface ApiOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  headers?: Record<string, string>
  body?: any
  timeout?: number
  showNotifications?: boolean
  loadingMessage?: string
  successMessage?: string
  errorMessage?: string
}

class ApiClient {
  private baseURL: string
  private defaultTimeout: number

  constructor(baseURL: string = '/api', timeout: number = 10000) {
    this.baseURL = baseURL
    this.defaultTimeout = timeout
  }

  private async request<T>(
    endpoint: string,
    options: ApiOptions = {}
  ): Promise<ApiResponse<T>> {
    const {
      method = 'GET',
      headers = {},
      body,
      timeout = this.defaultTimeout,
      showNotifications = true,
      loadingMessage,
      successMessage,
      errorMessage
    } = options

    const notificationStore = useNotificationStore()
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeout)

    try {
      // Show loading notification if requested
      if (showNotifications && loadingMessage) {
        notificationStore.info('Loading', loadingMessage, { duration: 0 })
      }

      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...headers
        },
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()

      // Show success notification if requested
      if (showNotifications && successMessage) {
        notificationStore.success('Success', successMessage)
      }

      return {
        data,
        status: response.status,
        message: 'Success'
      }
    } catch (error) {
      clearTimeout(timeoutId)
      
      let errorMessage = 'An unexpected error occurred'
      
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          errorMessage = 'Request timed out'
        } else {
          errorMessage = error.message
        }
      }

      // Show error notification if requested
      if (showNotifications) {
        notificationStore.error(
          'Error',
          errorMessage || errorMessage,
          { duration: 8000 }
        )
      }

      throw {
        message: errorMessage,
        status: 500,
        code: 'UNKNOWN_ERROR'
      } as ApiError
    }
  }

  // Convenience methods
  async get<T>(endpoint: string, options?: Omit<ApiOptions, 'method'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'GET' })
  }

  async post<T>(endpoint: string, body?: any, options?: Omit<ApiOptions, 'method' | 'body'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'POST', body })
  }

  async put<T>(endpoint: string, body?: any, options?: Omit<ApiOptions, 'method' | 'body'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'PUT', body })
  }

  async patch<T>(endpoint: string, body?: any, options?: Omit<ApiOptions, 'method' | 'body'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'PATCH', body })
  }

  async delete<T>(endpoint: string, options?: Omit<ApiOptions, 'method'>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' })
  }
}

// Create default instance
export const api = new ApiClient()

// Export the class for custom instances
export { ApiClient }

// Utility functions for common API patterns
export const withRetry = async <T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> => {
  let lastError: Error

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error as Error
      
      if (i < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)))
      }
    }
  }

  throw lastError!
}

export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: ReturnType<typeof setTimeout>
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
} 