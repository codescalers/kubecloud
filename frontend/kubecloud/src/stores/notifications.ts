import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  duration?: number
  action?: {
    text: string
    handler: () => void
  }
}

export const useNotificationStore = defineStore('notifications', () => {
  const notifications = ref<Notification[]>([])
  const timeouts: Record<string, ReturnType<typeof setTimeout>> = {}

  const addNotification = (notification: Omit<Notification, 'id'>) => {
    const id = Date.now().toString() + Math.random().toString(36).substr(2, 9)
    const duration = notification.duration || 5000
    const newNotification: Notification = {
      ...notification,
      id,
      duration
    }
    
    notifications.value.push(newNotification)
    
    // Auto-remove after duration
    if (duration > 0) {
      const timeout = setTimeout(() => {
        removeNotification(id)
      }, duration)
      timeouts[id] = timeout
    }
    
    return id
  }

  const removeNotification = (id: string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
    
    // Clear timeout
    const timeout = timeouts[id]
    if (timeout) {
      clearTimeout(timeout)
      delete timeouts[id]
    }
  }

  const clearAllNotifications = () => {
    notifications.value = []
    Object.values(timeouts).forEach(timeout => clearTimeout(timeout))
    Object.keys(timeouts).forEach(key => delete timeouts[key])
  }

  // Convenience methods for different notification types
  const success = (title: string, message: string, options?: Partial<Omit<Notification, 'id' | 'type' | 'title' | 'message'>>) => {
    return addNotification({
      type: 'success',
      title,
      message,
      ...options
    })
  }

  const error = (title: string, message: string, options?: Partial<Omit<Notification, 'id' | 'type' | 'title' | 'message'>>) => {
    return addNotification({
      type: 'error',
      title,
      message,
      ...options
    })
  }

  const warning = (title: string, message: string, options?: Partial<Omit<Notification, 'id' | 'type' | 'title' | 'message'>>) => {
    return addNotification({
      type: 'warning',
      title,
      message,
      ...options
    })
  }

  const info = (title: string, message: string, options?: Partial<Omit<Notification, 'id' | 'type' | 'title' | 'message'>>) => {
    return addNotification({
      type: 'info',
      title,
      message,
      ...options
    })
  }

  return {
    notifications,
    addNotification,
    removeNotification,
    clearAllNotifications,
    success,
    error,
    warning,
    info
  }
}) 