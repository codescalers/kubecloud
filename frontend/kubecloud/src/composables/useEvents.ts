import { useUserStore } from "@/stores/user"
import type { DeploymentEvent } from "./useDeploymentEvents"
import { ref, onMounted, onUnmounted, watch, readonly } from 'vue'
export interface EventListenerOptions {
  endpoint?: string
  token?: string
  withCredentials?: boolean
  maxReconnectAttempts?: number
  reconnectDelay?: number
  onMessage?: (event: DeploymentEvent) => void
  onConnect?: () => void
  onError?: (error: Event) => void
}


export function useGenericEventListener(options: EventListenerOptions = {}) {
  const eventSource = ref<EventSource | null>(null)
  const userStore = useUserStore()

  const isConnected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = options.maxReconnectAttempts ?? 5
  const reconnectDelay = options.reconnectDelay ?? 2000

  function connect() {
    console.log('Connecting to SSE')
    if (eventSource.value) return

    const backendBaseUrl = (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_API_BASE_URL) || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
    const token = options.token || userStore.token || ''
    const endpoint = options.endpoint || '/v1/events'
    const url = backendBaseUrl + endpoint + '?token=' + encodeURIComponent(token)

    eventSource.value = new EventSource(url, {
      withCredentials: options.withCredentials ?? true
    })

    eventSource.value.onopen = () => {
      isConnected.value = true
      reconnectAttempts.value = 0
      console.log('SSE connection established')
      options.onConnect?.()
    }

    eventSource.value.onmessage = (event) => {
      const data = JSON.parse(event.data) as DeploymentEvent
      const type = data.type || 'info'

      if (type === 'connected') {
        isConnected.value = true
        console.log('SSE connected')
        return
      }

      options.onMessage?.(data)
    }

    eventSource.value.onerror = (err) => {
      isConnected.value = false
      console.error('SSE connection error:', err)

      options.onError?.(err)

      // Attempt to reconnect
      if (reconnectAttempts.value < maxReconnectAttempts) {
        setTimeout(() => {
          reconnectAttempts.value++
          disconnect()
          connect()
        }, reconnectDelay * reconnectAttempts.value)
      }
    }
  }

  function disconnect() {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
    isConnected.value = false
  }

  return {
    connect,
    disconnect,
    isConnected: readonly(isConnected),
    reconnectAttempts: readonly(reconnectAttempts)
  }
}
