import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService, type LoginRequest, type RegisterRequest } from '@/utils/authService'
import { api, createWorkflowStatusChecker } from '@/utils/api'
import type { ApiResponse, VerifyCodeRequest } from '@/utils/authService'
import { userService } from '@/utils/userService'
import { useNotificationStore } from './notifications'
import { WorkflowStatus } from '@/types/ewf'
import router from '@/router'

export interface User {
  id: number
  username: string
  email: string
  admin: boolean
  verified: boolean
  updated_at: string
  balance_usd?: number
  pending_balance_usd?: number
  balance: number
}

export interface AuthState {
  user: User | null
  token: string | null
  isLoading: boolean
  error: string | null
}

export const useUserStore = defineStore('user',
  // Store definition
  () => {
    // State
    const user = ref<User | null>(null)
    const token = ref<string | null>(null)
    const isLoading = ref(false)
    const error = ref<string | null>(null)
    const netBalance = ref(0)
    const pendingBalance = ref(0)
    const balanceInterval = ref<ReturnType<typeof setInterval> | null>(null)

    // Computed properties
    const isAdmin = computed(() => user.value?.admin)
    const isLoggedIn = computed(() => !!token.value)

    // Actions
    const startBalanceRefresh = () => {
      if (balanceInterval.value) return
      balanceInterval.value = setInterval(() => {
        updateNetBalance()
      }, 30000) // Refresh every 30 seconds
    }
    const stopBalanceRefresh = () => {
      if (balanceInterval.value) {
        clearInterval(balanceInterval.value)
        balanceInterval.value = null
      }
    }

    const loadUser = async () => {
      const userRes = await api.get<ApiResponse<{ user: User }>>('/v1/user/', { requiresAuth: true, showNotifications: false })
      user.value = userRes.data.data.user
    }

    const login = async (email: string, password: string) => {
      isLoading.value = true
      error.value = null

      try {
        const loginData: LoginRequest = { email, password }
        const response = await authService.login(loginData)
        authService.storeTokens(response.access_token, response.refresh_token)
        token.value = response.access_token
        await loadUser()
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Login failed'
        throw err
      } finally {
        isLoading.value = false
      }
    }

    const logout = () => {
      user.value = null
      token.value = null
      error.value = null
      // Clear localStorage
      authService.clearTokens()
      // Clear notifications and related ephemeral state
      useNotificationStore().reset()
    }

    interface RegisterFormData {
      name: string
      email: string
      password: string
      confirmPassword: string
    }

    const register = async (formData: RegisterFormData) => {
      isLoading.value = true
      error.value = null

      try {
        const registerData: RegisterRequest = {
          name: formData.name,
          email: formData.email,
          password: formData.password,
          confirm_password: formData.confirmPassword,
        }
        const response = await authService.register(registerData)
        return response
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Registration failed'
        throw err
      } finally {
        isLoading.value = false
      }
    }

    const verifyCode = async (data: VerifyCodeRequest) => {
      isLoading.value = true
      error.value = null
      try {
        const response = await authService.verifyCode(data)
        authService.storeTokens(response.access_token, response.refresh_token)
        token.value = response.access_token
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Verification failed'
        throw err
      } finally {
        isLoading.value = false
      }
    }

    const updateProfile = async (updates: Partial<User>) => {
      if (!user.value) throw new Error('User not logged in')

      isLoading.value = true
      error.value = null

      try {
        // TODO: Implement profile update API call
        user.value = { ...user.value, ...updates }
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Profile update failed'
        throw err
      } finally {
        isLoading.value = false
      }
    }

    const refreshToken = async () => {
      const tokens = authService.getTokens()
      if (!tokens.refreshToken) return

      try {
        const response = await authService.refreshToken({ refresh_token: tokens.refreshToken })
        // Only access_token is returned in RefreshTokenResponse, so keep the old refreshToken
        authService.storeTokens(response.access_token, tokens.refreshToken)
        token.value = response.access_token
      } catch (err) {
        logout()
        throw {
          message: 'Token refresh failed',
          silent: true
        }
      }
    }

    const initializeAuth = () => {
      // Only set token if it exists in localStorage
      const { accessToken } = authService.getTokens()
      if (accessToken) {
        token.value = accessToken
      }
    }

    const updateNetBalance = async () => {
      const balance = await userService.fetchBalance()
      netBalance.value = balance.balance
      pendingBalance.value = balance.pending_balance
    }

    return {
      // State (raw refs for persistence)
      user,
      token,
      isLoading,
      error,
      netBalance,
      pendingBalance,

      // Computed
      isAdmin,
      isLoggedIn,

      // Actions
      login,
      logout,
      register,
      verifyCode,
      updateProfile,
      refreshToken,
      initializeAuth,
      updateNetBalance,
      startBalanceRefresh,
      stopBalanceRefresh,
      loadUser,
    }
  },
  // Persisted state options
  {
    persist: {
      pick: ['user', 'token']
    }
  }
)
