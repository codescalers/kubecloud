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
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%),
    radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
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
  border-radius: 1.5rem;
  box-shadow: 0 4px 24px 0 rgba(59,130,246,0.10);
  padding: 2.5rem 2rem 2rem 2rem;
  z-index: 2;
  border: 1px solid rgba(96, 165, 250, 0.15);
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: fadeInUp 0.7s cubic-bezier(0.4,0,0.2,1);
}
.auth-header {
  text-align: center;
  margin-bottom: 2rem;
}
.auth-title {
  font-size: clamp(2rem, 4vw, 2.5rem);
  font-weight: 500;
  color: #fff;
  letter-spacing: -0.5px;
  line-height: 1.1;
}
.auth-subtitle {
  font-size: clamp(1rem, 2vw, 1.2rem);
  color: #60a5fa;
  opacity: 0.92;
  font-weight: 400;
}
.auth-form {
  width: 100%;
}
.auth-options-vertical {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}
.auth-submit-btn {
  font-size: 1rem;
  padding: 0.9rem 0;
  border-radius: 1.5rem;
  font-weight: 500;
  margin-top: 0.2rem;
  box-shadow: 0 4px 24px 0 rgba(59,130,246,0.10);
  background: transparent;
  color: #fff;
  border: 2px solid #fff;
  transition: box-shadow 0.2s, transform 0.2s, background 0.2s, color 0.2s;
}
.auth-submit-btn:hover {
  background: #fff;
  color: #0a192f;
  box-shadow: 0 8px 32px 0 rgba(59,130,246,0.18);
  transform: translateY(-2px) scale(1.04);
}
.auth-footer {
  text-align: center;
  margin-top: 1.5rem;
}
.auth-footer-text {
  color: #CBD5E1;
  margin-right: 0.5rem;
}
.auth-link {
  font-weight: 500;
  text-transform: none;
  letter-spacing: 0.01em;
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
    padding: 1.2rem 0.5rem 1.5rem 0.5rem;
    min-width: 0;
    max-width: 98vw;
  }
  .auth-title {
    font-size: clamp(2rem, 8vw, 3rem);
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
