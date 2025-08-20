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
    return response.data.data.users
  }

  // Delete a user (requires admin auth)
  async deleteUser(userId: number): Promise<DeleteUserResponse> {
    const response = await api.delete<DeleteUserResponse>(`/v1/users/${userId}`, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Deleting user...',
      successMessage: 'User deleted successfully',
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
      successMessage: 'User credited successfully',
      errorMessage: 'Failed to credit user'
    })
    return response.data
  }

  // Generate vouchers (requires admin auth)
  async generateVouchers(data: GenerateVouchersRequest): Promise<GenerateVouchersResponse> {
    const response = await api.post<GenerateVouchersResponse>('/v1/vouchers/generate', data, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Generating vouchers...',
      successMessage: 'Vouchers generated successfully',
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
    return response.data.data.vouchers
  }

  // List all invoices (requires admin auth)
  async listInvoices(): Promise<Invoice[]> {
    const response = await api.get<ApiResponse<{ invoices: Invoice[] }>>('/v1/invoices', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load invoices'
    })
    return response.data.data.invoices
  }

      // List all pending records (requires admin auth)
  async listPendingRecords(): Promise<PendingRecord[]> {
    const response = await api.get<ApiResponse<{ pending_records: PendingRecord[] }>>('/v1/pending-records', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load payments'
    })
    return response.data.data.pending_records
  }

  // Send a system email to all users (requires admin auth)
  async sendSystemEmail(formData: FormData): Promise<SystemEmailResponse> {
    const response = await api.post<SystemEmailResponse>('/v1/users/mail', formData, {
      requiresAuth: true,
      showNotifications: true,
      loadingMessage: 'Sending email to all users',
      successMessage: 'Email sent to all users',
      errorMessage: 'Failed to send email',
      contentType: '',
      timeout: 60000,
    })
    return response.data
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
