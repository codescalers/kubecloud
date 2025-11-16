import { api } from './api'

export interface SystemStats {
  total_users: number
  total_clusters: number
  up_nodes: number
  countries: number
  cores: number
  ssd: number
  system_account_balance: number
}

export class StatsService {
  private static instance: StatsService

  private constructor() {}

  static getInstance(): StatsService {
    if (!StatsService.instance) {
      StatsService.instance = new StatsService()
    }
    return StatsService.instance
  }

  // Get system statistics
  async getStats(): Promise<SystemStats> {
    const response = await api.get<{data: SystemStats}>('/v1/stats', {
      requiresAuth: false,
      showNotifications: false,
      errorMessage: 'Failed to load system statistics'
    })
    return response.data.data || {
      total_users: 0,
      total_clusters: 0,
      up_nodes: 0,
      countries: 0,
      cores: 0,
      ssd: 0,
      system_account_balance: 0
    }
  }
}

export const statsService = StatsService.getInstance()
