import { formatDistanceToNow } from 'date-fns'
import type { NotificationType, NotificationSeverity } from '../types/notifications'

/**
 * Get notification icon for different types
 * @param type - Notification type
 * @returns Icon name
 */
export function getNotificationIcon(type: NotificationType): string {
  switch (type) {
    case 'deployment': return 'mdi-rocket-launch'
    case 'billing': return 'mdi-credit-card'
    case 'user': return 'mdi-account'
    case 'connected': return 'mdi-link'
    case 'node': return 'mdi-server'
    default: return 'mdi-bell'
  }
}

/**
 * Get notification color for different types
 * @param type - Notification type
 * @returns Color name
 */
export function getNotificationColor(type: NotificationType): string {
  switch (type) {
    case 'deployment': return 'success'
    case 'billing': return 'warning'
    case 'user': return 'info'
    case 'connected': return 'primary'
    case 'node': return 'info'
    default: return 'grey'
  }
}

/**
 * Get toast notification icon for different severities
 * @param severity - Notification severity
 * @returns Icon name
 */
export function getToastIcon(severity: NotificationSeverity): string {
  switch (severity) {
    case 'success': return 'mdi-check-circle'
    case 'error': return 'mdi-alert-circle'
    case 'warning': return 'mdi-alert'
    case 'info': return 'mdi-information'
    default: return 'mdi-information'
  }
}

/**
 * Get toast notification color for different severities
 * @param severity - Notification severity
 * @returns Color hex value
 */
export function getToastColor(severity: NotificationSeverity): string {
  switch (severity) {
    case 'success': return '#10B981'
    case 'error': return '#EF4444'
    case 'warning': return '#F59E0B'
    case 'info': return '#60a5fa'
    default: return '#60a5fa'
  }
}

/**
 * Format notification timestamp to relative time
 * @param timestamp - ISO timestamp string
 * @returns Formatted relative time string
 */
export function formatNotificationTime(timestamp: string): string {
  try {
    return formatDistanceToNow(new Date(timestamp), { addSuffix: true })
  } catch {
    return 'Unknown time'
  }
}
