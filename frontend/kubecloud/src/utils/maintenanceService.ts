import { api, type ApiResponse } from './api'

export interface MaintenanceStatus {
  enabled: boolean
}

export class MaintenanceService {
  /**
   * Fetch the current maintenance status from the public endpoint
   */
  static async getMaintenanceStatus(): Promise<MaintenanceStatus> {
    try {
      const response = await api.get<ApiResponse<MaintenanceStatus>>('/v1/system/maintenance/status', {
        requiresAuth: false,
        showNotifications: false,
        timeout: 5000
      })
      return response.data.data
    } catch (error) {
      console.error('Failed to fetch maintenance status:', error)
      // Return default status if API fails
      return { enabled: false }
    }
  }

  /**
   * Check if a route is allowed during maintenance mode
   */
  static isRouteAllowedDuringMaintenance(routePath: string): boolean {
    const allowedRoutes = [
      '/',           // home
      '/features',   // features
      '/use-cases',  // usecases
      '/docs',       // docs
      '/nodes'       // reserve (public view)
    ]
    
    return allowedRoutes.includes(routePath)
  }
}
