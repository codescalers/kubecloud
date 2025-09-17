import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'
import type { NotificationSeverity, NotificationStatus, BaseNotification, Notification } from '../types/notifications'
import { useUserStore } from './user'

export const useNotificationStore = defineStore('notifications', () => {
  // State
  const notifications = ref<Notification[]>([])
  const loading = ref(false)
  const activeTimeouts = ref<Map<string, NodeJS.Timeout>>(new Map())
  
  // Computed
  const unreadCount = computed(() => {
    return notifications.value.filter(n => n.persistent && n.status === 'unread').length
  })
  
  const toastNotifications = computed(() => 
    notifications.value.filter(n => !n.persistent)
  )
  
  const persistentNotifications = computed(() => 
    notifications.value.filter(n => n.persistent)
  )

  // Core functions
  const addNotification = (notification: Omit<Notification, 'id'>) => {
    const id = notification.persistent ? `temp-${Date.now()}` : `toast-${Date.now()}-${Math.random()}`
    const newNotification: Notification = {
      ...notification,
      id,
      persistent: notification.persistent || false,
      duration: notification.duration || (notification.persistent ? undefined : 5000)
    }
    
    notifications.value.unshift(newNotification)
    
    // Auto-remove toast notifications
    if (!notification.persistent && newNotification.duration) {
      const timeout = setTimeout(() => {
        removeNotification(String(id))
        activeTimeouts.value.delete(String(id))
      }, newNotification.duration)
      activeTimeouts.value.set(String(id), timeout)
    }
    
    return id
  }

  const removeNotification = (id: number | string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      // Clear timeout if exists
      const timeout = activeTimeouts.value.get(String(id))
      if (timeout) {
        clearTimeout(timeout)
        activeTimeouts.value.delete(String(id))
      }
      notifications.value.splice(index, 1)
    }
  }

  const setReadStatus = async (id: string, target: NotificationStatus) => {
    const notification = notifications.value.find(n => n.id === id)
    if (!notification || !notification.persistent) return
    if (notification.status === target) return
    try {
      const action = target === 'read' ? 'read' : 'unread'
      await api.patch(`/v1/notifications/${id}/${action}`, undefined, { requiresAuth: true })
      notification.status = target
      notification.read_at = target === 'read' ? new Date().toISOString() : undefined
    } catch (error) {
      console.error(`Failed to mark notification as ${target}:`, error)
      throw error
    }
  }

  const markAsRead = async (id: string) => setReadStatus(id, 'read')

  const markAsUnread = async (id: string) => setReadStatus(id, 'unread')

  const markAllAsRead = async () => {
    const unreadNotifications = notifications.value.filter(n => n.persistent && n.status === 'unread')
    if (unreadNotifications.length === 0) return

    try {
      await api.patch('/v1/notifications/read-all', undefined, { requiresAuth: true })
      unreadNotifications.forEach(notification => {
        notification.status = 'read'
        notification.read_at = new Date().toISOString()
      })
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
      throw error
    }
  }

  const clearAll = async () => {
    const persistentNotifications = notifications.value.filter(n => n.persistent)
    if (persistentNotifications.length === 0) return

    try {
      await api.delete('/v1/notifications', { requiresAuth: true })
      notifications.value = notifications.value.filter(n => !n.persistent)
    } catch (error) {
      console.error('Failed to clear all notifications:', error)
    }
  }

  const deleteNotification = async (id: string) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification && notification.persistent) {
      try {
        await api.delete(`/v1/notifications/${id}`, { requiresAuth: true })
        removeNotification(id)
      } catch (error) {
        console.error('Failed to delete notification:', error)
        throw error
      }
    }
  }

  // Internal: replace persistent notifications
  const replacePersistent = (serverNotifications: BaseNotification[]) => {
    notifications.value = serverNotifications.map(n => ({ ...n, persistent: true }))
  }

  // Internal: fetch endpoint and replace
  const fetchAndReplace = async (endpoint: string) => {
    const userStore = useUserStore()
    if (!userStore.token) return
    if (loading.value) return
    try {
      loading.value = true
      const response = await api.get(endpoint, { requiresAuth: true })
      const list = (response as any)?.data?.data?.notifications as BaseNotification[] | undefined
      if (Array.isArray(list)) replacePersistent(list)
    } catch (error) {
      console.error('Failed to load notifications:', error)
    } finally {
      loading.value = false
    }
  }

  // Load persistent notifications from server
  const loadNotifications = async () => fetchAndReplace('/v1/notifications?limit=50')

  // Load unread notifications from server
  const loadUnreadNotifications = async () => fetchAndReplace('/v1/notifications/unread?limit=50')

  // Convenience methods for toast notifications
  // Internal: create toast
  const addToast = (severity: NotificationSeverity, title: string, message: string) =>
    addNotification({
      type: 'user',
      severity,
      payload: { title, message },
      status: 'read',
      persistent: false,
      created_at: new Date().toISOString()
    })

  const success = (title: string, message: string) => addToast('success', title, message)
  const error = (title: string, message: string) => addToast('error', title, message)
  const warning = (title: string, message: string) => addToast('warning', title, message)
  const info = (title: string, message: string) => addToast('info', title, message)

  // Cleanup function to prevent memory leaks
  const cleanup = () => {
    activeTimeouts.value.forEach(timeout => clearTimeout(timeout))
    activeTimeouts.value.clear()
  }

  // Reset store state (used on logout)
  const reset = () => {
    cleanup()
    notifications.value = []
    loading.value = false
  }

  return {
    // State
    notifications,
    loading,
    unreadCount,
    toastNotifications,
    persistentNotifications,
    
    // Actions
    addNotification,
    removeNotification,
    markAsRead,
    markAsUnread,
    markAllAsRead,
    clearAll,
    deleteNotification,
    loadNotifications,
    loadUnreadNotifications,
    
    // Toast convenience methods
    success,
    error,
    warning,
    info,
    
    // Cleanup
    cleanup,
    reset
  }
})