import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService, type LoginRequest, type RegisterRequest } from '@/utils/authService'
import { api } from '@/utils/api'
import type { ApiResponse } from '@/utils/authService'
import { userService } from '@/utils/userService'

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

    // Computed properties
    const isAdmin = computed(() => user.value?.admin)
    const isLoggedIn = computed(() => !!token.value)

    // Actions
    const login = async (email: string, password: string) => {
      isLoading.value = true
      error.value = null

      try {
        const loginData: LoginRequest = { email, password }
        const response = await authService.login(loginData)
        
        // Store tokens
        authService.storeTokens(response.access_token, response.refresh_token)
        
        // Set token in store
        token.value = response.access_token
        const userRes = await api.get<ApiResponse<{ user: User }>>('/v1/user/', { requiresAuth: true, showNotifications: false })
        user.value = userRes.data.data.user
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
          confirm_password: formData.confirmPassword
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
        // If refresh fails, logout user
        logout()
        throw err
      }
    }

    const initializeAuth = () => {
      // Only set token if it exists in localStorage
      const { accessToken } = authService.getTokens()
      if (accessToken) {
        token.value = accessToken
        // Optionally, fetch user profile here if you want to populate user.value on app start
        // (async () => {
        //   const userRes = await api.get<ApiResponse<{ user: User }>>('/v1/user/', { requiresAuth: true, showNotifications: false })
        //   user.value = userRes.data.data.user
        // })()
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
      updateProfile,
      refreshToken,
      initializeAuth,
      updateNetBalance,
    }
  },
  // Persisted state options
  {
    persist: {
      pick: ['user', 'token']
    }
  }
) 