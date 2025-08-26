import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'
import { useClusterStore } from '../stores/clusters'
import { useNodeManagement } from './useNodeManagement'

export interface DeploymentEvent {
  type: string
  data: any
  message?: string
  timestamp: string
}

export function useDeploymentEvents() {
  const eventSource = ref<EventSource | null>(null)
  const notificationStore = useNotificationStore()
  const userStore = useUserStore()
  const clusterStore = useClusterStore()
  const { fetchRentedNodes } = useNodeManagement()

  const isConnected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5
  const reconnectDelay = 2000 // 2 seconds

  function connect() {
    if (eventSource.value || isConnected.value) return
    const backendBaseUrl = (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_API_BASE_URL) || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
    const token = userStore.token || ''
    const url = backendBaseUrl + '/v1/events?token=' + encodeURIComponent(token)

    eventSource.value = new EventSource(url, { withCredentials: true })

    eventSource.value.onopen = () => {
      isConnected.value = true
      reconnectAttempts.value = 0
      console.log('SSE connection established')
    }

    eventSource.value.onmessage = (event) => {
        const data = JSON.parse(event.data) as DeploymentEvent
        const type = data.type || 'info'

        if (type === 'connected') {
          isConnected.value = true
          return
        }

        // Handle workflow updates
        if (type === 'workflow_update') {
          const message = data.message || data.data?.message
          if (message) {
            // Check if workflow failed based on message content
            if (message.toLowerCase().includes('failed')) {
              notificationStore.error('Workflow', message)
            } else if (message.toLowerCase().includes('completed')) {
              notificationStore.success('Workflow', message)
            }
          }

          // Always refresh data when workflow completes (success or failure)
          refreshClusterData()
        }
        else {
          notificationStore.info('', data.data)
        }
    }

    eventSource.value.onerror = (err) => {
      isConnected.value = false
      console.error('SSE connection error:', err)

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

  // Refresh all cluster-related data
  async function refreshClusterData() {
    await Promise.all([
      clusterStore.fetchClusters(),
      fetchRentedNodes()
    ])
  }

  // Watch for token changes to reconnect
  watch(() => userStore.token, (newToken) => {
    if (newToken && !isConnected.value) {
      connect()
    }
    if (!newToken && isConnected.value) {
      disconnect()
    }
  }, { immediate: true })

  onMounted(() => {
    // Simple fallback: if token exists but not connected after a short delay, connect
    setTimeout(() => {
      if (userStore.token && !isConnected.value) {
        connect()
      }
    }, 100)
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    connect,
    disconnect,
    refreshClusterData
  }
}
