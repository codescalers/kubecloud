import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { mockApi, mockUsers, MOCK_CONFIG } from '@/utils/mockData'

export interface User {
  id: number
  name: string
  email: string
  role: 'admin' | 'user' | 'developer'
  avatar?: string
  createdAt: string
  lastLogin: string
}

export interface AuthState {
  user: User | null
  token: string | null
  isLoading: boolean
  error: string | null
}

export const useUserStore = defineStore('user', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed properties
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isLoggedIn = computed(() => !!user.value && !!token.value)
  const isDeveloper = computed(() => user.value?.role === 'developer')

  // Actions
  const login = async (email: string, password: string) => {
    isLoading.value = true
    error.value = null

    try {
      if (MOCK_CONFIG.enabled) {
        // Mock login
        await mockApi.post('/auth/login', { email, password })
        
        // Find user by email
        const mockUser = mockUsers.find(u => u.email === email)
        if (!mockUser) {
          throw new Error('Invalid credentials')
        }

        user.value = mockUser
        token.value = `mock-token-${Date.now()}`
        
        // Store in localStorage for persistence
        localStorage.setItem('user', JSON.stringify(mockUser))
        localStorage.setItem('token', token.value)
      } else {
        // Real API call would go here
        // const response = await api.post('/auth/login', { email, password })
        // user.value = response.data.user
        // token.value = response.data.token
        throw new Error('API not ready - use mock mode')
      }
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
    localStorage.removeItem('user')
    localStorage.removeItem('token')
  }

  const register = async (userData: Omit<User, 'id' | 'createdAt' | 'lastLogin'>) => {
    isLoading.value = true
    error.value = null

    try {
      if (MOCK_CONFIG.enabled) {
        // Mock registration
        await mockApi.post('/auth/register', userData)
        
        const newUser: User = {
          ...userData,
          id: Date.now(),
          createdAt: new Date().toISOString(),
          lastLogin: new Date().toISOString(),
        }

        user.value = newUser
        token.value = `mock-token-${Date.now()}`
        
        // Store in localStorage
        localStorage.setItem('user', JSON.stringify(newUser))
        localStorage.setItem('token', token.value)
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
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
      if (MOCK_CONFIG.enabled) {
        // Mock profile update
        await mockApi.put(`/users/${user.value.id}`, updates)
        
        user.value = { ...user.value, ...updates }
        localStorage.setItem('user', JSON.stringify(user.value))
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Profile update failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const refreshToken = async () => {
    if (!token.value) return

    try {
      if (MOCK_CONFIG.enabled) {
        // Mock token refresh
        await mockApi.post('/auth/refresh', { token: token.value })
        token.value = `mock-token-${Date.now()}`
        localStorage.setItem('token', token.value)
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      // If refresh fails, logout user
      logout()
      throw err
    }
  }

  const initializeAuth = () => {
    // Check for stored auth data on app start
    const storedUser = localStorage.getItem('user')
    const storedToken = localStorage.getItem('token')

    if (storedUser && storedToken) {
      try {
        user.value = JSON.parse(storedUser)
        token.value = storedToken
      } catch (err) {
        // Clear invalid stored data
        localStorage.removeItem('user')
        localStorage.removeItem('token')
      }
    }
  }

  // Convenience methods for development
  const loginAsAdmin = () => {
    const adminUser = mockUsers.find(u => u.role === 'admin')
    if (adminUser) {
      user.value = adminUser
      token.value = `mock-token-admin-${Date.now()}`
      localStorage.setItem('user', JSON.stringify(adminUser))
      localStorage.setItem('token', token.value)
    }
  }

  const loginAsUser = () => {
    const regularUser = mockUsers.find(u => u.role === 'user')
    if (regularUser) {
      user.value = regularUser
      token.value = `mock-token-user-${Date.now()}`
      localStorage.setItem('user', JSON.stringify(regularUser))
      localStorage.setItem('token', token.value)
    }
  }

  return {
    // State
    user: computed(() => user.value),
    token: computed(() => token.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),

    // Computed
    isAdmin,
    isLoggedIn,
    isDeveloper,

    // Actions
    login,
    logout,
    register,
    updateProfile,
    refreshToken,
    initializeAuth,

    // Development helpers
    loginAsAdmin,
    loginAsUser,
  }
}) 