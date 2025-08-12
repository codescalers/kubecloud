import { WorkflowStatus } from '@/types/ewf'
import { api, createWorkflowStatusChecker } from './api'
import { useNotificationStore } from '@/stores/notifications'

// Types for auth requests and responses
export interface RegisterRequest {
  name: string
  email: string
  password: string
  confirm_password: string
}

export interface RegisterResponse {
  email: string
  workflow_id: string
}

export interface VerifyCodeRequest {
  email: string
  code: number
}

export interface VerifyCodeResponse {
  email: string
  workflow_id: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: {
    id: number
    username: string
    email: string
    admin: boolean
    verified: boolean
    updated_at: string
    credit_card_balance: number
    credited_balance: number
  }
}

// New type to match backend response
export interface BackendLoginResponse {
  message: string
  status: number
  data: LoginResponse
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface RefreshTokenResponse {
  access_token: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ForgotPasswordResponse {
  email: string
  timeout: string
}

export interface ChangePasswordRequest {
  email: string
  password: string
  confirm_password: string
}

export interface ChangePasswordResponse {
  message: string
}

// Generic API response type
export interface ApiResponse<T> {
  status: number;
  message?: string;
  data: T;
  error?: string;
}

// Auth service class
export class AuthService {
  private static instance: AuthService

  private constructor() {}

  static getInstance(): AuthService {
    if (!AuthService.instance) {
      AuthService.instance = new AuthService()
    }
    return AuthService.instance
  }

  // Register a new user
  async register(data: RegisterRequest): Promise<void> {
    const response = await api.post<ApiResponse<RegisterResponse>>('/v1/user/register', data, {
      showNotifications: true,
      loadingMessage: 'Creating your account...',
      errorMessage: 'Registration failed',
      timeout: 60000
    })
    const workflowChecker = createWorkflowStatusChecker(response.data.data.workflow_id, { initialDelay: 15000,interval: 3000 })
    const status = await workflowChecker.status
    if (status === WorkflowStatus.StatusCompleted) {
      useNotificationStore().success(
        'Registration Success',
        'User registered successfully',
      )
    }
    if (status === WorkflowStatus.StatusFailed) {
      useNotificationStore().error(
        'Registration Failed',
        'Failed to register user',
      )
      throw new Error('Failed to register user')
    }

  }

  // Verify registration code
  async verifyCode(data: VerifyCodeRequest): Promise<void> {
    const response = await api.post<ApiResponse<VerifyCodeResponse>>('/v1/user/register/verify', data, {
      showNotifications: true,
      errorMessage: 'Verification failed'
    })
    const workflowChecker = createWorkflowStatusChecker(response.data.data.workflow_id, { initialDelay: 3000, interval: 2000 })
    const status = await workflowChecker.status
    if (status === WorkflowStatus.StatusCompleted) {
      useNotificationStore().success(
        'Verification Success',
        'User verified successfully',
      )
    }
    if (status === WorkflowStatus.StatusFailed) {
      useNotificationStore().error(
        'Verification Failed',
        'Failed to verify user',
      )
      throw new Error('Failed to verify user')
    }
  }

  // Login user
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/v1/user/login', data, {
      showNotifications: true,
      successMessage: 'Welcome back!',
      errorMessage: 'Login failed'
    })
    return response.data.data
  }

  // Refresh access token
  async refreshToken(data: RefreshTokenRequest): Promise<RefreshTokenResponse> {
    const response = await api.post<ApiResponse<RefreshTokenResponse>>('/v1/user/refresh', data, {
      showNotifications: false // Don't show notifications for token refresh
    })
    return response.data.data
  }

  // Forgot password
  async forgotPassword(data: ForgotPasswordRequest): Promise<ForgotPasswordResponse> {
    const response = await api.post<ApiResponse<ForgotPasswordResponse>>('/v1/user/forgot_password', data, {
      showNotifications: true,
      loadingMessage: 'Sending reset code...',
      successMessage: 'Reset code sent to your email!',
      errorMessage: 'Failed to send reset code'
    })
    return response.data.data
  }

  // Verify forgot password code
  async verifyForgotPasswordCode(data: VerifyCodeRequest): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/v1/user/forgot_password/verify', data, {
      showNotifications: false,
      errorMessage: 'Invalid reset code'
    })
    return response.data.data
  }

  // Change password (requires authentication)
  async changePassword(data: ChangePasswordRequest, useTemporaryToken = false): Promise<ChangePasswordResponse> {
    let customToken: string | undefined

    if (useTemporaryToken) {
      const tempTokens = this.getTempTokens()
      customToken = tempTokens.accessToken || undefined
    }

    const response = await api.put<ApiResponse<ChangePasswordResponse>>('/v1/user/change_password', data, {
      requiresAuth: true,
      customToken,
      showNotifications: true,
      loadingMessage: 'Updating password...',
      successMessage: 'Password updated successfully!',
      errorMessage: 'Failed to update password'
    })
    return response.data.data
  }



  // Store tokens in localStorage
  storeTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem('token', accessToken)
    localStorage.setItem('refreshToken', refreshToken)
  }

  // Store temporary tokens for password reset (separate from main auth)
  storeTempTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem('temp_reset_token', accessToken)
    localStorage.setItem('temp_reset_refresh_token', refreshToken)
    // Set expiration time (15 minutes from now)
    const expirationTime = Date.now() + (15 * 60 * 1000)
    localStorage.setItem('temp_reset_token_expires', expirationTime.toString())
  }

  // Get temporary tokens for password reset
  getTempTokens(): { accessToken: string | null; refreshToken: string | null } {
    const token = localStorage.getItem('temp_reset_token')
    const refreshToken = localStorage.getItem('temp_reset_refresh_token')
    const expirationTime = localStorage.getItem('temp_reset_token_expires')

    // Check if tokens are expired
    if (token && expirationTime && Date.now() > parseInt(expirationTime)) {
      // Tokens expired, clean up
      this.clearTempTokens()
      return { accessToken: null, refreshToken: null }
    }

    return {
      accessToken: token,
      refreshToken: refreshToken
    }
  }

  // Clear temporary tokens
  clearTempTokens(): void {
    localStorage.removeItem('temp_reset_token')
    localStorage.removeItem('temp_reset_refresh_token')
    localStorage.removeItem('temp_reset_token_expires')
  }

  // Check if password reset session is valid
  isPasswordResetSessionValid(): boolean {
    const hasSession = localStorage.getItem('password_reset_session') === 'true'
    const tempTokens = this.getTempTokens()
    return hasSession && !!tempTokens.accessToken
  }



  // Get stored tokens
  getTokens(): { accessToken: string | null; refreshToken: string | null } {
    return {
      accessToken: localStorage.getItem('token'),
      refreshToken: localStorage.getItem('refreshToken')
    }
  }

  // Clear stored tokens
  clearTokens(): void {
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  // Clear all auth-related localStorage items
  clearAllAuthData(): void {
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('password_reset_session')
    localStorage.removeItem('user')
    this.clearTempTokens()
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!localStorage.getItem('token')
  }
}

// Export singleton instance
export const authService = AuthService.getInstance()
