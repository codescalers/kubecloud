import router from "@/router"
import { api } from "./api"
import type { ApiResponse } from "./authService"
import type { PendingRecord } from "./userService"

// Types for admin requests and responses
export interface User {
  id: number
  username: string
  email: string
  admin: boolean
  verified: boolean
  balance: number
  created_at: string
  updated_at: string
}

export interface Voucher {
  id: number
  voucher: string
  value: number
  used: boolean
  used_by?: number
  created_at: string
  expires_at: string
}

export interface GenerateVouchersRequest {
  count: number
  value: number
  expire_after_days: number
}

export interface GenerateVouchersResponse {
  message: string
  vouchers: Voucher[]
}

export interface CreditUserRequest {
  amount: number
  memo: string
}

export interface CreditUserResponse {
  message: string
  user: string
  amount: number
  memo: string
}

export interface DeleteUserResponse {
  message: string
}

export interface SystemEmail {
  title: string
  message: string
  priority: string
}

export interface SystemEmailResponse {
  failed_emails: string[],
  failed_emails_count: number,
  successful_emails: number,
  total_users: number
}

export interface AdminWorkflow {
  uuid: string
  name: string
  display_name: string
  status: string
  current_step: number
  total_steps: number
  step_name: string
  state: Record<string, any>
  user_id: number
  created_at: string
  queue_name: string
  metadata: Record<string, string>
}

export interface Invoice {
  id: number
  user_id: number
  total: number
  nodes: any[]
  tax: number
  created_at: string
}

// Admin service class
export class AdminService {
  private static instance: AdminService

  private constructor() {}

  static getInstance(): AdminService {
    if (!AdminService.instance) {
      AdminService.instance = new AdminService()
    }
    return AdminService.instance
  }

  // List all users (requires admin auth)
  async listUsers(): Promise<User[]> {
    const response = await api.get<ApiResponse<{ users: User[] }>>('/v1/users', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load users'
    })
    return response.data.data?.users || []
  }

  // Delete a user (requires admin auth)
  async deleteUser(userId: number): Promise<DeleteUserResponse> {
    const response = await api.delete<DeleteUserResponse>(`/v1/users/${userId}`, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Deleting user...',
      errorMessage: 'Failed to delete user'
    })
    return response.data
  }

  // Credit a user's balance (requires admin auth)
  async creditUser(userId: number, data: CreditUserRequest): Promise<CreditUserResponse> {
    const response = await api.post<CreditUserResponse>(`/v1/users/${userId}/credit`, data, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Crediting user...',
      errorMessage: 'Failed to credit user'
    })
    return response.data
  }

  // Drain a user's balance to system account (requires admin auth)
  async drainUser(userId: number): Promise<void> {
    await api.post<void>(`/v1/users/${userId}/drain`, {}, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Draining user balance...',
      errorMessage: 'Failed to drain user balance'
    })
  }

  // Drain all users' balances to system account (requires admin auth)
  async drainAllUsers(): Promise<void> {
    await api.post<void>(`/v1/users/drain-all`, {}, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Draining all users\' balances...',
      errorMessage: 'Failed to drain all users\' balances'
    })
  }

  // Generate vouchers (requires admin auth)
  async generateVouchers(data: GenerateVouchersRequest): Promise<GenerateVouchersResponse> {
    const response = await api.post<GenerateVouchersResponse>('/v1/vouchers/generate', data, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Generating vouchers...',
      errorMessage: 'Failed to generate vouchers'
    })
    return response.data
  }

  // List all vouchers (requires admin auth)
  async listVouchers(): Promise<Voucher[]> {
    const response = await api.get<ApiResponse<{ vouchers: Voucher[] }>>('/v1/vouchers', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load vouchers'
    })
    return response.data.data?.vouchers || []
  }

  // List all invoices (requires admin auth)
  async listInvoices(): Promise<Invoice[]> {
    const response = await api.get<ApiResponse<{ invoices: Invoice[] }>>('/v1/invoices', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load invoices'
    })
    return response.data.data?.invoices || []
  }

      // List all pending records (requires admin auth)
  async listPendingRecords(): Promise<PendingRecord[]> {
    const response = await api.get<ApiResponse<{ pending_records: PendingRecord[] }>>('/v1/pending-records', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load payments'
    })
    return response.data.data?.pending_records || []
  }

  // Send a system email to all users (requires admin auth)
  async sendSystemEmail(formData: FormData): Promise<SystemEmailResponse> {
    const response = await api.post<SystemEmailResponse>('/v1/users/mail', formData, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Sending email to all users',
      errorMessage: 'Failed to send email',
      contentType: '',
      timeout: 60000,
    })
    return response.data
  }

  // List all workflows (requires admin auth)
  async listWorkflows(status?: string): Promise<AdminWorkflow[]> {
    const url = status ? `/v1/workflows?status=${status}` : '/v1/workflows'
    const response = await api.get<ApiResponse<{ workflows: AdminWorkflow[] }>>(url, {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load workflows'
    })
    return response.data.data?.workflows || []
  }

  // List workflows with pagination
  async listWorkflowsPaginated(status?: string, page: number = 1, limit: number = 10): Promise<{ workflows: AdminWorkflow[]; total: number; page: number; limit: number; total_pages: number }> {
    try {
      const params = new URLSearchParams()
      if (status) params.append('status', status)
      params.append('page', page.toString())
      params.append('limit', limit.toString())
      
      const url = `/v1/workflows?${params.toString()}`
      const response = await api.get<ApiResponse<{ workflows: AdminWorkflow[]; total: number; page: number; limit: number; total_pages: number }>>(url, {
        requiresAuth: true,
        errorMessage: 'Failed to load workflows'
      })

      // Handle the response safely
      if (!response || !response.data) {
        console.warn('Empty response from workflows endpoint')
        return { workflows: [], total: 0, page, limit, total_pages: 0 }
      }

      const data = response.data.data
      if (!data) {
        console.warn('No data in workflows response')
        return { workflows: [], total: 0, page, limit, total_pages: 0 }
      }

      return {
        workflows: data.workflows || [],
        total: data.total || 0,
        page: data.page || page,
        limit: data.limit || limit,
        total_pages: data.total_pages || 0
      }
    } catch (error) {
      console.error('Error fetching workflows:', error)
      throw error
    }
  }


  async SetMaintenanceModeStatus(status: boolean): Promise<void> {
    try {
      const response = await api.put('/v1/system/maintenance/status', { enabled: status }, {
        requiresAuth: true,
        showNotifications: true,
        loadingMessage: 'Setting maintenance mode...',
        successMessage: 'Maintenance mode set successfully, redirecting to maintenance page in 3 seconds',
        errorMessage: 'Failed to set maintenance mode'
      })
      setTimeout(() => {
        router.push('/maintenance')
      }, 3000)
    } catch (error) {
      console.error(error)
      throw error
    }
  }
}

export const adminService = AdminService.getInstance()
