import type { User } from "../types/user"

/**
 * Format a date string to a readable format
 * @param dateStr - ISO date string
 * @returns Formatted date string
 */
export function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * Format a date string to date only
 * @param dateStr - ISO date string
 * @returns Formatted date string
 */
export function formatDateOnly(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString()
}

/**
 * Format a date string to time only
 * @param dateStr - ISO date string
 * @returns Formatted time string
 */
export function formatTimeOnly(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function calculateNetBalance(user: User) {
  return user.credit_card_balance_in_usd + user.credited_balance_in_usd
}
