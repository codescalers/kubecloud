import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'
import { useClusterStore } from '../stores/clusters'
import { useNodeManagement } from './useNodeManagement'

// TypeScript interfaces matching the backend notification model
export type NotificationType = 'deployment' | 'billing' | 'user' | 'connected'
export type NotificationSeverity = 'info' | 'error' | 'warning' | 'success'

export interface SSEMessage {
  type: NotificationType
  data: any
  severity: NotificationSeverity
  task_id?: string
  timestamp: string
}

export function useNotificationEvents() {
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

    const backendBaseUrl =
      (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_API_BASE_URL) ||
      import.meta.env.VITE_API_BASE_URL ||
      'http://localhost:8080/api'
    const token = userStore.token || ''
    const url = backendBaseUrl + '/v1/events?token=' + encodeURIComponent(token)

    eventSource.value = new EventSource(url, { withCredentials: true })

    eventSource.value.onopen = () => {
      isConnected.value = true
      reconnectAttempts.value = 0
      console.log('Notification SSE connection established')
    }

    eventSource.value.onmessage = (event) => {
      try {
        const eventData = JSON.parse(event.data) as SSEMessage
        handleSSEMessage(eventData)
      } catch (error) {
        console.error('Error parsing SSE message:', error)
      }
    }

    eventSource.value.onerror = (err) => {
      isConnected.value = false
      console.error('Notification SSE connection error:', err)

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

  function handleSSEMessage(message: SSEMessage) {
    const { type, data, severity } = message

    // Handle connection confirmation
    if (type === 'connected') {
      isConnected.value = true
      return
    }

    // Get title and message content
    const title = getNotificationTitle(type)
    const messageText = getNotificationMessage(message)

    // Display notification based on severity
    switch (severity) {
      case 'success':
        notificationStore.success(title, messageText)
        break
      case 'error':
        notificationStore.error(title, messageText)
        break
      case 'warning':
        notificationStore.warning(title, messageText)
        break
      case 'info':
      default:
        notificationStore.info(title, messageText)
        break
    }

    // Handle specific notification types
    handleSpecificNotificationType(type, message)
  }

  function getNotificationTitle(type: string): string {
    switch (type as NotificationType) {
      case 'deployment':
        return 'Deployment'
      case 'billing':
        return 'Billing'
      case 'user':
        return 'Account'
      case 'connected':
        return 'Connection'
      default:
        return 'System'
    }
  }

  function getNotificationMessage(message: SSEMessage): string {
    const { data, type } = message

    // Handle different data formats
    if (typeof data === 'string') {
      return data
    }

    if (data && typeof data === 'object') {
      // Try to get message from data object
      if (data.message) {
        return data.message
      }

      if (data.title && data.description) {
        return `${data.title}: ${data.description}`
      }

      if (data.title) {
        return data.title
      }

      if (data.description) {
        return data.description
      }

      if (data.status) {
        return `Status: ${data.status}`
      }
    }

    // Fallback based on type
    switch (type as NotificationType) {
      case 'deployment':
        return 'Deployment status updated'
      case 'billing':
        return 'Billing information updated'
      case 'user':
        return 'Account information updated'
      case 'connected':
        return 'Connected to notification service'
      default:
        return 'System notification'
    }
  }

  function handleSpecificNotificationType(type: string, message: SSEMessage) {
    switch (type as NotificationType) {
      case 'deployment':
        handleDeploymentNotification(message)
        break
      case 'billing':
        handleBillingNotification(message)
        break
      case 'user':
        handleUserNotification(message)
        break
      case 'connected':
        // Connection notifications don't need special handling
        break
    }
  }

  function handleDeploymentNotification(message: SSEMessage) {
    refreshClusterData()
  }

  function handleBillingNotification(message: SSEMessage) {
    // TODO: Handle billing-specific logic if needed
    console.log('Billing notification received:', message)
  }

  function handleUserNotification(message: SSEMessage) {
    // TODO: Handle user-specific logic if needed
    console.log('User notification received:', message)
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
    try {
      await Promise.all([clusterStore.fetchClusters(), fetchRentedNodes()])
    } catch (error) {
      console.error('Error refreshing cluster data:', error)
    }
  }

  // Watch for token changes to reconnect
  watch(
    () => userStore.token,
    (newToken) => {
      if (newToken && !isConnected.value) {
        connect()
      }
      if (!newToken && isConnected.value) {
        disconnect()
      }
    },
    { immediate: true },
  )

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
    isConnected,
    refreshClusterData,
  }
}
