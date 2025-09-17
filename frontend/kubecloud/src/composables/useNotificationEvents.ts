import { ref, onMounted, onUnmounted, watch, watchEffect } from 'vue'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'
import { useClusterStore } from '../stores/clusters'
import { useNodeManagement } from './useNodeManagement'
import type { NotificationType, NotificationSeverity } from '@/types/notifications'
import router from '@/router'

/** Core notification data structure */
interface NotificationData {
  /** Notification title/subject */
  subject: string
  /** Detailed notification message */
  message: string
}

/** Server-Sent Event message structure from backend */
interface SSEMessage {
  /** Type of notification event */
  type: NotificationType
  /** Notification content data */
  data: NotificationData
  /** Severity level for UI styling */
  severity: NotificationSeverity
  /** Optional task identifier for tracking */
  task_id?: string
  /** Event timestamp */
  timestamp: string
}

const MAX_RECONNECT_ATTEMPTS = 5
const RECONNECT_DELAY = 2000
const NOTIFICATION_DELAY = 2000
const MAX_NOTIFICATIONS_QUEUE = 10

/**
 * Vue composable for managing real-time notification events via Server-Sent Events (SSE)
 *
 * Provides a reactive connection to the backend notification system with automatic
 * reconnection, message handling, and integration with various stores for data updates.
 *
 * @returns Object containing connection methods and reactive state
 */
export function useNotificationEvents() {
  const eventSource = ref<EventSource | null>(null)
  const notificationStore = useNotificationStore()
  const userStore = useUserStore()
  const clusterStore = useClusterStore()
  const { fetchRentedNodes } = useNodeManagement()
  const notificationQueue = ref<SSEMessage[]>([])
  const processingQueue = ref(false)
  const isConnected = ref(false)
  const reconnectAttempts = ref(0)
  const isOnline = ref(navigator.onLine)
  const isPageVisible = ref(document.visibilityState === 'visible')
  const shouldReconnectOnVisibility = ref(false)
  const eventListenersInitialized = ref(false)

  /**
   * Establishes SSE connection to the backend notification service
   *
   * Creates an EventSource connection with authentication token and sets up
   * event handlers for open, message, and error events. Includes automatic
   * reconnection logic on connection failures.
   */
  function connect() {
    console.log('[SSE Debug] Attempting to connect to SSE')
    if (eventSource.value || isConnected.value || !userStore.token || !isOnline.value) return

    const backendBaseUrl =
      (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_API_BASE_URL) ||
      import.meta.env.VITE_API_BASE_URL ||
      'http://localhost:8080/api'
    const token = userStore.token
    const url = backendBaseUrl + '/v1/events?token=' + encodeURIComponent(token)

    eventSource.value = new EventSource(url, { withCredentials: true })

    eventSource.value.onopen = () => {
      isConnected.value = true
      reconnectAttempts.value = 0
      shouldReconnectOnVisibility.value = false
      console.log('[SSE] Notification SSE connection established successfully')
    }

    eventSource.value.onmessage = (event) => {
      try {
        const eventData = JSON.parse(event.data) as SSEMessage
        // Remove the oldest notification if the queue is full to prevent memory overflow
        if (notificationQueue.value.length >= MAX_NOTIFICATIONS_QUEUE) {
          notificationQueue.value.shift()
        }
        notificationQueue.value.push(eventData)
      } catch (error) {
        console.error('[SSE] Error parsing SSE message:', error, 'Raw data:', event.data)
      }
    }

    eventSource.value.onerror = (err) => {
      isConnected.value = false
      console.error('[SSE Debug] Notification SSE connection error:', err)

      // Only attempt reconnection if we're online and the page is visible
      if (
        isOnline.value &&
        isPageVisible.value &&
        reconnectAttempts.value < MAX_RECONNECT_ATTEMPTS
      ) {
        const delay = RECONNECT_DELAY * 2 ** reconnectAttempts.value
        setTimeout(async () => {
          reconnectAttempts.value++
          disconnect().then(connect)
        }, delay)
      }
      if (!isPageVisible.value) {
        shouldReconnectOnVisibility.value = true
      }
      if (!isOnline.value) {
        console.log('[SSE] Device offline, will reconnect when network is restored')
      }
    }
  }

  async function processNotificationQueue() {
    if (processingQueue.value) return
    processingQueue.value = true
    while (notificationQueue.value.length > 0) {
      const event = notificationQueue.value.shift()!
      handleSSEMessage(event)
      await new Promise((resolve) => setTimeout(resolve, NOTIFICATION_DELAY))
    }
    processingQueue.value = false
  }

  watchEffect(() => {
    if (notificationQueue.value.length > 0) {
      processNotificationQueue()
    }
  })

  /**
   * Processes incoming SSE messages and routes them to appropriate handlers
   *
   * Parses the message data, displays notifications via the notification store,
   * and triggers specific actions based on the notification type.
   *
   * @param event The SSE message containing notification data
   */
  function handleSSEMessage(event: SSEMessage) {
    const { type, data, severity } = event

    if (type === 'connected') {
      isConnected.value = true
      return
    }

    const { subject, message } = getNotificationData(data, type)

    switch (severity) {
      case 'success':
        notificationStore.success(subject, message)
        handleSpecificNotificationType(type)
        break
      case 'error':
        notificationStore.error(subject, message)
        break
      case 'warning':
        notificationStore.warning(subject, message)
        break
      case 'info':
      default:
        notificationStore.info(subject, message)
        break
    }
  }

  /**
   * Extracts and formats notification subject and message from event data
   *
   * Ensures both subject and message are present, falling back to defaults
   * based on notification type if not provided in the data.
   *
   * @param data The notification data from the SSE event
   * @param type The type of notification for fallback generation
   * @returns Formatted subject and message strings
   */
  function getNotificationData(
    data: NotificationData,
    type: NotificationType,
  ): { subject: string; message: string } {
    let message = ''
    let subject = ''
    if (data?.message) {
      message = data.message
    }
    if (data?.subject) {
      subject = data.subject
    }

    if (!message) message = parseDefaultNotificationMessage(data, type)
    if (!subject) {
      subject = type.charAt(0).toUpperCase() + type.slice(1)
    }
    return { subject, message }
  }

  /**
   * Generates default notification messages based on notification type
   *
   * Provides fallback messages when the notification data doesn't include
   * a specific message, ensuring users always receive meaningful feedback.
   *
   * @param data The notification data (may be string or object)
   * @param type The notification type for message generation
   * @returns Default message string for the notification type
   */
  function parseDefaultNotificationMessage(data: NotificationData, type: NotificationType): string {
    if (typeof data === 'string') {
      return data
    }

    // Fallback based on type
    switch (type as NotificationType) {
      case 'deployment':
        return 'Deployment status update'
      case 'billing':
        return 'Billing information update'
      case 'user':
        return 'Account information update'
      case 'connected':
        return 'Connected to notification service'
      case 'node':
        return 'Node status update'
      default:
        return 'System notification'
    }
  }

  /**
   * Routes notification types to their specific handler functions
   *
   * Triggers appropriate actions based on notification type, such as
   * refreshing data stores or updating UI state.
   *
   * @param type The notification type to handle
   */
  function handleSpecificNotificationType(type: NotificationType) {
    switch (type) {
      case 'deployment':
        handleDeploymentNotification()
        break
      case 'node':
        handleNodeNotification()
        break
      case 'billing':
        handleBillingNotification()
        break
      case 'user':
        handleUserNotification()
        break
      case 'connected':
        break
    }
  }

  /**
   * Handles deployment-related notifications
   *
   * Refreshes cluster data to reflect deployment status changes.
   */
  function handleDeploymentNotification() {
    refreshClusterData()
  }

  /**
   * Handles node-related notifications
   *
   * Fetches updated node rental information when node status changes.
   */
  function handleNodeNotification() {
    fetchRentedNodes()
  }

  /**
   * Handles billing-related notifications
   *
   * Updates user's net balance when billing information changes.
   */
  function handleBillingNotification() {
    userStore.updateNetBalance()
  }

  /**
   * Handles user-related notifications
   *
   * Currently logs the notification; can be extended for user-specific logic.
   */
  async function handleUserNotification() {
    console.log('User notification received')
    await userStore.loadUser()
    router.push('/dashboard')
  }

  /**
   * Closes the SSE connection
   *
   * Properly terminates the EventSource connection and updates connection state.
   */
  function disconnect(): Promise<void> {
    return new Promise((resolve) => {
      if (eventSource.value) {
        eventSource.value.close()
        eventSource.value = null
      }
      isConnected.value = false
      resolve()
    })
  }

  /**
   * Refreshes all cluster-related data from the backend
   *
   * Fetches updated cluster information and rented nodes data in parallel
   * to ensure UI reflects the latest state after deployment notifications.
   */
  async function refreshClusterData() {
    try {
      await clusterStore.fetchClusters()
    } catch (error) {
      console.error('Error refreshing cluster data:', error)
    }
  }

  /**
   * Handles network status changes
   */
  function handleOnlineStatusChange() {
    isOnline.value = navigator.onLine
    if (isOnline.value && userStore.token && !isConnected.value) {
      reconnectAttempts.value = 0
      connect()
    }
  }

  /**
   * Handles page visibility changes
   */
  function handleVisibilityChange() {
    isPageVisible.value = document.visibilityState === 'visible'
    if (
      isPageVisible.value &&
      shouldReconnectOnVisibility.value &&
      userStore.token &&
      !isConnected.value
    ) {
      shouldReconnectOnVisibility.value = false
      reconnectAttempts.value = 0
      connect()
    } else if (!isPageVisible.value && isConnected.value) {
      // Mark that we should reconnect when page becomes visible again
      shouldReconnectOnVisibility.value = true
    }
  }

  /**
   * Sets up event listeners for network and visibility changes
   */
  function setupNetworkAndVisibilityListeners() {
    if (eventListenersInitialized.value) return
    if (typeof window === 'undefined' || typeof document === 'undefined') return
    window.addEventListener('online', handleOnlineStatusChange)
    window.addEventListener('offline', handleOnlineStatusChange)
    document.addEventListener('visibilitychange', handleVisibilityChange)
    eventListenersInitialized.value = true
  }

  /**
   * Removes event listeners for network and visibility changes
   */
  function removeNetworkAndVisibilityListeners() {
    if (!eventListenersInitialized.value) return
    window.removeEventListener('online', handleOnlineStatusChange)
    window.removeEventListener('offline', handleOnlineStatusChange)
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    eventListenersInitialized.value = false
  }

  async function cleanup() {
    removeNetworkAndVisibilityListeners()
    notificationQueue.value = []
    processingQueue.value = false
    shouldReconnectOnVisibility.value = false
    reconnectAttempts.value = 0
    isConnected.value = false
  }

  onMounted(() => {
    setupNetworkAndVisibilityListeners()
  })

  onUnmounted(async () => {
    await disconnect()
    cleanup()
  })

  /**
   * Watches for token changes to reconnect
   *
   * Automatically reconnects to the SSE connection when the user's authentication token changes.
   * Disconnects when the token is removed.
   */
  watch(
    () => userStore.token,
    async (newToken) => {
      await disconnect()
      if (newToken) connect()
    },
    { immediate: true },
  )

  return {
    isConnected,
  }
}
