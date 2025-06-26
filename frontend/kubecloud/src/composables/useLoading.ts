import { ref, computed, readonly } from 'vue'

export interface LoadingState {
  isLoading: boolean
  progress: number
  message: string
  error: string | null
}

export function useLoading(initialMessage = 'Loading...') {
  const isLoading = ref(false)
  const progress = ref(0)
  const message = ref(initialMessage)
  const error = ref<string | null>(null)

  const loadingState = computed<LoadingState>(() => ({
    isLoading: isLoading.value,
    progress: progress.value,
    message: message.value,
    error: error.value
  }))

  const start = (msg?: string) => {
    isLoading.value = true
    progress.value = 0
    message.value = msg || initialMessage
    error.value = null
  }

  const updateProgress = (newProgress: number, msg?: string) => {
    progress.value = Math.min(100, Math.max(0, newProgress))
    if (msg) {
      message.value = msg
    }
  }

  const setMessage = (msg: string) => {
    message.value = msg
  }

  const setError = (err: string) => {
    error.value = err
    isLoading.value = false
  }

  const complete = (msg?: string) => {
    progress.value = 100
    if (msg) {
      message.value = msg
    }
    setTimeout(() => {
      isLoading.value = false
      progress.value = 0
    }, 500)
  }

  const stop = () => {
    isLoading.value = false
    progress.value = 0
    error.value = null
  }

  const reset = () => {
    isLoading.value = false
    progress.value = 0
    message.value = initialMessage
    error.value = null
  }

  // Utility for wrapping async operations
  const withLoading = async <T>(
    operation: () => Promise<T>,
    loadingMsg?: string,
    successMsg?: string
  ): Promise<T> => {
    try {
      start(loadingMsg)
      const result = await operation()
      complete(successMsg)
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred'
      setError(errorMessage)
      throw err
    }
  }

  // Utility for operations with progress updates
  const withProgress = async <T>(
    operation: (updateProgress: (progress: number, message?: string) => void) => Promise<T>,
    loadingMsg?: string,
    successMsg?: string
  ): Promise<T> => {
    try {
      start(loadingMsg)
      const result = await operation(updateProgress)
      complete(successMsg)
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred'
      setError(errorMessage)
      throw err
    }
  }

  return {
    // State
    isLoading: readonly(isLoading),
    progress: readonly(progress),
    message: readonly(message),
    error: readonly(error),
    loadingState,

    // Methods
    start,
    updateProgress,
    setMessage,
    setError,
    complete,
    stop,
    reset,

    // Utilities
    withLoading,
    withProgress
  }
} 