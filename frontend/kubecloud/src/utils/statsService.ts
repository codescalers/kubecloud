import { api } from './api'

export interface SystemStats {
  total_users: number
  total_clusters: number
  up_nodes: number
  countries: number
  cores: number
  ssd: number
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
    const response = await api.get<SystemStats>('/v1/stats', {
      requiresAuth: false,
      showNotifications: false,
      errorMessage: 'Failed to load system statistics'
    })
    return response.data
  }
}

export const statsService = StatsService.getInstance()
