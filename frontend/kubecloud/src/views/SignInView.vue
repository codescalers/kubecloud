<template>
  <div class="auth-container">
    <div class="auth-background">
      <div class="auth-particles"></div>
    </div>

    <v-card class="auth-card card-enhanced">
      <div class="auth-header">
        <h1 class="auth-title kubecloud-gradient">Welcome Back!</h1>
        <p class="auth-subtitle">Sign in to your KubeCloud account</p>
      </div>

      <v-form @submit.prevent="handleSignIn" class="auth-form">
        <v-text-field
          v-model="form.email"
          label="Email Address"
          type="email"
          prepend-inner-icon="mdi-email"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.email"
          :disabled="loading"
          required
        />

        <v-text-field
          v-model="form.password"
          label="Password"
          type="password"
          prepend-inner-icon="mdi-lock"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.password"
          :disabled="loading"
          required
        />

        <div class="auth-options-vertical">
          <v-checkbox
            v-model="form.rememberMe"
            label="Remember me"
            color="primary"
            :disabled="loading"
          />
          <v-btn
            variant="text"
            size="small"
            class="forgot-password-btn kubecloud-hover-blue"
            :disabled="loading"
          >
            Forgot Password?
          </v-btn>
        </div>

        <v-btn
          type="submit"
          color="primary"
          block
          size="large"
          class="auth-submit-btn btn-enhanced kubecloud-glow-blue"
          :loading="loading"
          :disabled="loading"
        >
          <v-icon icon="mdi-login" class="mr-2"></v-icon>
          {{ loading ? 'Signing In...' : 'Sign In' }}
        </v-btn>
      </v-form>

      <div class="auth-footer">
        <span class="auth-footer-text">Don't have an account?</span>
        <v-btn
          variant="text"
          color="primary"
          to="/sign-up"
          class="auth-link kubecloud-hover-blue"
          :disabled="loading"
        >
          Sign Up
        </v-btn>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '../stores/notifications'
import { useLoading } from '../composables/useLoading'
import { validateForm, VALIDATION_RULES } from '../utils/validation'

const router = useRouter()
const notificationStore = useNotificationStore()
const { isLoading: loading, withLoading } = useLoading()

const form = reactive({
  email: '',
  password: '',
  rememberMe: false
})

const errors = reactive({
  email: '',
  password: ''
})

const clearErrors = () => {
  errors.email = ''
  errors.password = ''
}

const validateFormData = () => {
  clearErrors()

  const validationFields = {
    email: {
      value: form.email,
      rules: VALIDATION_RULES.EMAIL
    },
    password: {
      value: form.password,
      rules: { required: true, minLength: 1 }
    }
  }

  const result = validateForm(validationFields)

  if (!result.isValid) {
    result.errors.forEach(error => {
      if (error.includes('email')) {
        errors.email = error
      } else if (error.includes('password')) {
        errors.password = error
      }
    })
    return false
  }

  return true
}

const handleSignIn = async () => {
  if (!validateFormData()) {
    notificationStore.error('Validation Error', 'Please fix the errors above')
    return
  }

  try {
    await withLoading(
      async () => {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 2000))

        // Simulate success
        if (form.email === 'test@example.com' && form.password === 'password') {
          notificationStore.success('Welcome Back!', 'Successfully signed in to your account')
          router.push('/dashboard')
        } else {
          throw new Error('Invalid email or password')
        }
      },
      'Signing in...',
      'Successfully signed in!'
    )
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Sign in failed'
    notificationStore.error('Sign In Failed', message)
  }
}
</script>

<style scoped>
.auth-container {
  min-height: 100vh;
  background: var(--kubecloud-navy);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  position: relative;
  overflow: hidden;
}

.auth-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--color-background) !important;
}

.auth-particles {
  display: none !important;
}

.auth-card {
  min-width: 320px !important;
  max-width: 400px !important;
  width: 100%;
  padding: 2rem 1.5rem;
  background: var(--color-surface) !important;
  border: 1px solid var(--color-border) !important;
  border-radius: var(--rounded);
  box-shadow: none !important;
  color: var(--color-text);
  transition: border-color 0.15s;
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;
}

.auth-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--color-text);
  background: none;
}

.auth-subtitle {
  color: var(--color-text-muted);
  font-size: 1rem;
  margin-bottom: 0.5rem;
}

.auth-form {
  margin-bottom: 2rem;
}

.auth-field {
  margin-bottom: 0.75rem;
}

.auth-field :deep(.v-field) {
  background: var(--color-surface) !important;
  border: 1px solid var(--color-border) !important;
  border-radius: var(--rounded) !important;
  color: var(--color-text) !important;
  box-shadow: none !important;
  transition: border-color 0.15s;
}

.auth-field :deep(.v-field:hover),
.auth-field :deep(.v-field--focused) {
  border-color: var(--color-accent) !important;
  box-shadow: none !important;
}

.auth-options-vertical {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-bottom: 2rem;
}

.forgot-password-btn {
  margin-top: 0.15rem;
  font-size: 0.98rem;
  color: var(--color-accent);
  align-self: flex-start;
  padding-left: 2px;
  line-height: 1.2;
}

.auth-submit-btn {
  background: var(--color-accent) !important;
  color: #fff !important;
  font-weight: 600;
  font-size: 1rem;
  padding: 1rem;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none !important;
  transition: background 0.15s, color 0.15s;
}

.auth-submit-btn:hover,
.auth-submit-btn:focus {
  background: var(--color-accent) !important;
  color: #fff !important;
  box-shadow: none !important;
}

.auth-footer {
  text-align: center;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.5rem;
}

.auth-footer-text {
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.auth-link {
  font-weight: 600;
  color: #fff !important;
  text-decoration: underline;
}

@media (max-width: 600px) {
  .auth-container {
    padding: 1rem;
  }
  .auth-card {
    padding: 1.25rem 0.5rem;
    min-width: 0 !important;
    max-width: 100vw !important;
  }
  .auth-title {
    font-size: 1.2rem;
  }
}
</style>
