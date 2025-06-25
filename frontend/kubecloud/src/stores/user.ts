import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  // Mock user object
  const user = ref<null | { id: number; name: string; email: string; role: string }>(
    {
      id: 1,
      name: 'Admin User',
      email: 'admin@example.com',
      role: 'admin', // or 'user'
    }
  )

  // Computed property to check if user is admin
  const isAdmin = computed(() => user.value?.role === 'admin')

  // Computed property to check if user is logged in
  const isLoggedIn = computed(() => !!user.value)

  // Mock login/logout actions
  function loginAsAdmin() {
    user.value = {
      id: 1,
      name: 'Admin User',
      email: 'admin@example.com',
      role: 'admin',
    }
  }
  function loginAsUser() {
    user.value = {
      id: 2,
      name: 'Regular User',
      email: 'user@example.com',
      role: 'user',
    }
  }
  function logout() {
    user.value = null
  }

  return { user, isAdmin, isLoggedIn, loginAsAdmin, loginAsUser, logout }
}) 