export type NotificationType = 'deployment' | 'billing' | 'user' | 'connected' | 'node'
export type NotificationSeverity = 'info' | 'error' | 'warning' | 'success'
export type NotificationStatus = 'read' | 'unread'

// Base notification interface with common fields
export interface BaseNotification {
  id: string
  type: NotificationType
  severity: NotificationSeverity
  payload: Record<string, string>
  status: NotificationStatus
  created_at: string
  read_at?: string
  task_id?: string
}

export interface Notification extends BaseNotification {
  // Frontend-specific fields
  duration?: number
  persistent?: boolean
}

