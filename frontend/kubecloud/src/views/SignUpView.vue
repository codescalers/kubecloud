<template>
  <div class="auth-view">
    <LoadingComponent v-if="loading" fullPage message="Creating account..." />
    <div  class="auth-content fade-in">
      <div class="auth-header">
        <h1 class="auth-title">Create Account</h1>
        <p class="auth-subtitle">Join KubeCloud and start your journey</p>
      </div>
      <v-form @submit.prevent="handleSignUp" class="auth-form">
        <v-text-field
          v-model="form.name"
          label="Name"
          prepend-inner-icon="mdi-account"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.name"
          required
        />
        <v-text-field
          v-model="form.email"
          label="Email Address"
          type="email"
          prepend-inner-icon="mdi-email"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.email"
          required
        />
        <v-text-field
          v-model="form.password"
          label="Password"
          :type="showPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append-inner="showPassword = !showPassword"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.password"
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
          v-model="form.confirmPassword"
          label="Confirm Password"
          :type="showConfirmPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock-check"
          :append-inner-icon="showConfirmPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append-inner="showConfirmPassword = !showConfirmPassword"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.confirmPassword"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
        >
          <v-icon icon="mdi-account-plus" class="mr-2"></v-icon>
          {{ 'Create Account' }}
        </v-btn>
      </v-form>
      <div class="auth-footer">
        <span class="auth-footer-text">Already have an account?</span>
        <v-btn
          variant="outlined"
          color="white"
          to="/sign-in"
        >
          Sign In
        </v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'
import { validateForm, VALIDATION_RULES } from '../utils/validation'
import LoadingComponent from '../components/LoadingComponent.vue'

const router = useRouter()
const notificationStore = useNotificationStore()
const userStore = useUserStore()
const loading = ref(false)
const form = reactive({
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const errors = reactive({
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const showPassword = ref(false)
const showConfirmPassword = ref(false)

const clearErrors = () => {
  errors.name = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''
}

const validateFormData = () => {
  clearErrors()

  if (form.name.length < 3) {
    errors.name = 'Name must be at least 3 characters'
    return false
  }
  if (form.name.length > 64) {
    errors.name = 'Name must be no more than 64 characters'
    return false
  }

  const validationFields = {
    name: {
      value: form.name,
      rules: { required: true, minLength: 3 }
    },
    email: {
      value: form.email,
      rules: VALIDATION_RULES.EMAIL
    },
    password: {
      value: form.password,
      rules: VALIDATION_RULES.PASSWORD
    },
    confirmPassword: {
      value: form.confirmPassword,
      rules: {
        required: true,
        custom: (value: string) => {
          if (value !== form.password) {
            return 'Passwords do not match'
          }
          return true
        }
      }
    }
  }

  const result = validateForm(validationFields)

  if (!result.isValid) {
    result.errors.forEach(error => {
      if (error.includes('name')) {
        errors.name = error
      } else if (error.includes('email')) {
        errors.email = error
      } else if (error.includes('password') && !error.includes('confirm')) {
        errors.password = error
      } else if (error.includes('confirmPassword') || error.includes('do not match')) {
        errors.confirmPassword = error
      }
    })
    return false
  }
  return true
}

const handleSignUp = async () => {
  if (!validateFormData()) {
    // Don't show generic notification - inline errors are already shown
    return
  }
  loading.value = true
  try {
    await userStore.register({
      name: form.name,
      email: form.email,
      password: form.password,
      confirmPassword: form.confirmPassword
    })

    // Redirect to verify page on success
    router.push({ path: '/register/verify', query: { email: form.email } })
  } catch (error) {
    // Error handling is done in the auth service
    console.error('Sign up error:', error)
  } finally {
    loading.value = false
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
  background: rgba(30, 41, 59, 0.95);
  border-radius: 16px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
  padding: 2.5rem 2rem 2rem 2rem;
  z-index: 2;
  border: 2px solid white;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: fadeInUp 0.7s cubic-bezier(0.4,0,0.2,1);
  backdrop-filter: blur(12px);
}
.auth-header {
  text-align: center;
  margin-bottom: var(--space-8);
}
.auth-title {
  font-size: 1.875rem;
  font-weight: 700;
  color: #FFFFFF;
  letter-spacing: -0.5px;
  line-height: 1.1;
}
.auth-subtitle {
  font-size: 1.125rem;
  color: #F1F5F9;
  opacity: 0.92;
  font-weight: 400;
  margin-bottom: 0;
}
.auth-form {
  width: 100%;
}
.name-fields {
  display: flex;
  gap: var(--space-2);
}
@media (max-width: 600px) {
  .name-fields {
    flex-direction: column;
    gap: var(--space-1);
  }
}
.auth-options {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  margin-bottom: var(--space-4);
}
.v-btn[type="submit"] {
  font-size: 1rem;
  padding: 0.75rem 0;
  border-radius: 12px;
  font-weight: 500;
  margin-top: 0.5rem;
  border: 2px solid white !important;
  color: white !important;
}
.auth-footer {
  text-align: center;
  margin-top: 1.5rem;
}
.auth-footer-text {
  color: #CBD5E1;
  margin-right: 0.5rem;
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

/* Enhanced field styling for better visibility */
.auth-field :deep(.v-field) {
  border: 2px solid white !important;
  background: rgba(51, 65, 85, 0.8) !important;
}

.auth-field :deep(.v-field--focused) {
  border-color: #3B82F6 !important;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2) !important;
}

.auth-field :deep(.v-field__input) {
  color: white !important;
}

.auth-field :deep(.v-field__prepend-inner .v-icon) {
  color: #CBD5E1 !important;
}

.auth-field :deep(.v-field__append-inner) {
  cursor: pointer;
  transition: color 0.2s ease;
}

.auth-field :deep(.v-field__append-inner:hover) {
  color: #3B82F6 !important;
}

.auth-field :deep(.v-field__append-inner .v-icon) {
  font-size: 1.2rem;
  color: #CBD5E1 !important;
}

.auth-field :deep(.v-label) {
  color: #CBD5E1 !important;
}

.auth-field :deep(.v-field--focused .v-label) {
  color: #3B82F6 !important;
}

/* Password field styling */
.auth-field :deep(.v-field__append-inner) {
  cursor: pointer;
  transition: color 0.2s ease;
}

.auth-field :deep(.v-field__append-inner:hover) {
  color: var(--color-primary, #3B82F6);
}

.auth-field :deep(.v-field__append-inner .v-icon) {
  font-size: 1.2rem;
}

.password-requirements {
  margin-bottom: var(--space-4);
}

.text-muted {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.requirements-list {
  list-style-type: disc;
  padding-left: var(--space-4);
}
</style>
