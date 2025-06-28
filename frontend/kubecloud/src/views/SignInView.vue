<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content fade-in">
      <div class="auth-header">
        <h1 class="auth-title">Welcome Back!</h1>
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
          color="white"
          block
          size="large"
          variant="outlined"
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
          variant="outlined"
          color="white"
          to="/sign-up"
          :disabled="loading"
        >
          Sign Up
        </v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
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

onMounted(() => {
  const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px -50px 0px'
  }
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, observerOptions)
  document.querySelectorAll('.fade-in').forEach(el => {
    observer.observe(el)
  })
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
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 0;
  pointer-events: none;
}
.auth-content {
  min-width: 320px;
  max-width: 400px;
  width: 100%;
  background: rgba(10, 25, 47, 0.92);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  padding: var(--space-10) var(--space-8) var(--space-8) var(--space-8);
  z-index: 2;
  border: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: fadeInUp 0.7s cubic-bezier(0.4,0,0.2,1);
}
.auth-header {
  text-align: center;
  margin-bottom: var(--space-8);
}
.auth-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  letter-spacing: -0.5px;
  line-height: 1.1;
}
.auth-subtitle {
  font-size: var(--font-size-lg);
  color: #fff;
  opacity: 0.92;
  font-weight: var(--font-weight-normal);
}
.auth-form {
  width: 100%;
}
.auth-options-vertical {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
}
.auth-field {
  margin-bottom: var(--space-4);
}
.v-btn[type="submit"] {
  @apply btn btn-primary btn-full;
  font-size: var(--font-size-base);
  padding: var(--space-3) 0;
  border-radius: var(--radius-xl);
  font-weight: var(--font-weight-medium);
  margin-top: var(--space-2);
}
.auth-footer {
  text-align: center;
  margin-top: var(--space-6);
}
.auth-footer-text {
  color: var(--color-text-secondary);
  margin-right: var(--space-2);
}
.fade-in {
  opacity: 0;
  transform: translateY(40px);
  transition: opacity 0.7s cubic-bezier(0.4,0,0.2,1), transform 0.7s cubic-bezier(0.4,0,0.2,1);
}
.fade-in.visible {
  opacity: 1;
  transform: none;
}
@media (max-width: 600px) {
  .auth-content {
    padding: var(--space-6) var(--space-2) var(--space-4) var(--space-2);
    min-width: 0;
    max-width: 98vw;
  }
  .auth-title {
    font-size: var(--font-size-2xl);
  }
}
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(40px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}
</style>
