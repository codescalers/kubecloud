<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content">
      <div class="auth-header">
        <h1 class="auth-title">Reset Password</h1>
        <p class="auth-subtitle">Enter your new password below.</p>
      </div>
      <v-form @submit.prevent="handleResetPassword" class="auth-form">
        <v-text-field
          v-model="password"
          label="New Password"
          :type="showPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append-inner="showPassword = !showPassword"
          variant="outlined"
          class="auth-field"
          :error-messages="passwordError"
          :disabled="loading"
          required
        />
        <div class="password-requirements">
          <small class="text-muted">
            Password must contain at least 8 characters, including:
          </small>
          <ul class="requirements-list">
            <li>One uppercase letter (A-Z)</li>
            <li>One lowercase letter (a-z)</li>
            <li>One number (0-9)</li>
            <li>One special character (@$!%*?&)</li>
          </ul>
        </div>
        <v-text-field
          v-model="confirmPassword"
          label="Confirm Password"
          :type="showConfirmPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock-check"
          :append-inner-icon="showConfirmPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append-inner="showConfirmPassword = !showConfirmPassword"
          variant="outlined"
          class="auth-field"
          :error-messages="confirmPasswordError"
          :disabled="loading"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :loading="loading"
          :disabled="loading || !isFormValid"
        >
          <v-icon icon="mdi-check" class="mr-2"></v-icon>
          {{ loading ? 'Resetting...' : 'Reset Password' }}
        </v-btn>
      </v-form>
      <div class="auth-footer">
        <v-btn variant="text" color="white" to="/sign-in">Back to Sign In</v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { authService } from '@/utils/authService'

import { validateField, VALIDATION_RULES } from '../utils/validation'

const router = useRouter()
const route = useRoute()
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)

// Form validation using the same rules as signup
const passwordError = computed(() => {
  if (password.value.length === 0) return ''

  const validation = validateField({
    value: password.value,
    rules: VALIDATION_RULES.PASSWORD,
    fieldName: 'Password'
  })

  if (!validation.isValid) {
    return validation.errors[0]
  }

  return error.value
})

const confirmPasswordError = computed(() => {
  if (confirmPassword.value.length === 0) return ''

  const validation = validateField({
    value: confirmPassword.value,
    rules: {
      required: true,
      custom: (value: string) => {
        if (value !== password.value) {
          return 'Passwords do not match'
        }
        return true
      }
    },
    fieldName: 'Confirm Password'
  })

  if (!validation.isValid) {
    return validation.errors[0]
  }

  return ''
})

const isFormValid = computed(() => {
  const passwordValidation = validateField({
    value: password.value,
    rules: VALIDATION_RULES.PASSWORD
  })

  const confirmPasswordValidation = validateField({
    value: confirmPassword.value,
    rules: {
      required: true,
      custom: (value: string) => value === password.value
    }
  })

  return passwordValidation.isValid && confirmPasswordValidation.isValid
})

const getEmail = () => {
  return route.query.email as string || ''
}

const isPasswordResetSession = () => {
  return authService.isPasswordResetSessionValid()
}

const handleResetPassword = async () => {
  if (!isFormValid.value) return

  error.value = ''
  loading.value = true
  try {
    if (!isPasswordResetSession()) {
      throw new Error('Invalid password reset session')
    }

    await authService.changePassword({
      email: getEmail(),
      password: password.value,
      confirm_password: confirmPassword.value
    }, true) // Use temporary token

    // Clear all auth data including password reset session
    authService.clearAllAuthData()

    // Redirect to sign-in page
    router.push('/sign-in')
  } catch (err: any) {
    error.value = err?.message || 'Failed to reset password'
  } finally {
    loading.value = false
  }
}

// Guard against direct access without proper reset flow
onMounted(() => {
  const hasEmail = !!getEmail()
  const hasValidSession = isPasswordResetSession()

  // If user doesn't have email or valid reset session, redirect to forgot password
  if (!hasEmail || !hasValidSession) {
    router.push('/forgot-password')
  }
})


</script>

<style scoped>
.auth-view {
  min-height: 100vh;
  width: 100vw;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
}
.auth-background {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 0;
  pointer-events: none;
}
.auth-content {
  position: relative;
  z-index: 1;
  background: rgba(10, 25, 47, 0.95);
  border-radius: 2rem;
  box-shadow: 0 8px 32px 0 rgba(16, 42, 67, 0.25);
  padding: 3rem 2.5rem 2.5rem 2.5rem;
  max-width: 400px;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  animation: fadeInUp 0.6s ease-out;
}
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
.auth-header { text-align: center; }
.auth-title { font-size: 2.2rem; font-weight: 600; color: #fff; margin-bottom: 0.5rem; }
.auth-subtitle { color: #fff; font-size: 1.1rem; }
.auth-form { width: 100%; display: flex; flex-direction: column; gap: 1.5rem; }
.auth-field { width: 100%; }
.auth-footer { margin-top: 2rem; text-align: center; }
.password-requirements {
  margin-bottom: 1.5rem;
}
.text-muted {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.875rem;
}
.requirements-list {
  list-style-type: disc;
  padding-left: 1.5rem;
  margin-top: 0.5rem;
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.875rem;
}
.requirements-list li {
  margin-bottom: 0.25rem;
}
</style>
