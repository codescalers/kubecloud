<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content fade-in">
      <div class="auth-header">
        <h1 class="auth-title">Create Account</h1>
        <p class="auth-subtitle">Join KubeCloud and start your journey</p>
      </div>
      <v-form @submit.prevent="handleSignUp" class="auth-form">
        <div class="name-fields">
          <v-text-field
            v-model="form.firstName"
            label="First Name"
            prepend-inner-icon="mdi-account"
            variant="outlined"
            class="auth-field"
            :error-messages="errors.firstName"
            :disabled="loading"
            required
          />
          <v-text-field
            v-model="form.lastName"
            label="Last Name"
            prepend-inner-icon="mdi-account"
            variant="outlined"
            class="auth-field"
            :error-messages="errors.lastName"
            :disabled="loading"
            required
          />
        </div>
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
        <v-text-field
          v-model="form.confirmPassword"
          label="Confirm Password"
          type="password"
          prepend-inner-icon="mdi-lock-check"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.confirmPassword"
          :disabled="loading"
          required
        />
        <div class="auth-options">
          <v-checkbox
            v-model="form.agreeToTerms"
            label="I agree to the Terms of Service and Privacy Policy"
            color="primary"
            :disabled="loading"
            class="terms-checkbox"
          />
          <v-checkbox
            v-model="form.newsletter"
            label="Subscribe to our newsletter for updates"
            color="primary"
            :disabled="loading"
          />
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
          <v-icon icon="mdi-account-plus" class="mr-2"></v-icon>
          {{ loading ? 'Creating Account...' : 'Create Account' }}
        </v-btn>
      </v-form>
      <div class="auth-footer">
        <span class="auth-footer-text">Already have an account?</span>
        <v-btn
          variant="outlined"
          color="white"
          to="/sign-in"
          :disabled="loading"
        >
          Sign In
        </v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '../stores/notifications'
import { useLoading } from '../composables/useLoading'
import { validateForm, VALIDATION_RULES } from '../utils/validation'

const router = useRouter()
const notificationStore = useNotificationStore()
const { isLoading: loading, withLoading } = useLoading()

const form = reactive({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  confirmPassword: '',
  agreeToTerms: false,
  newsletter: false
})

const errors = reactive({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const clearErrors = () => {
  errors.firstName = ''
  errors.lastName = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''
}

const validateFormData = () => {
  clearErrors()

  const validationFields = {
    firstName: {
      value: form.firstName,
      rules: { required: true, minLength: 2 }
    },
    lastName: {
      value: form.lastName,
      rules: { required: true, minLength: 2 }
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
      if (error.includes('firstName')) {
        errors.firstName = error
      } else if (error.includes('lastName')) {
        errors.lastName = error
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

  if (!form.agreeToTerms) {
    notificationStore.error('Terms Required', 'You must agree to the Terms of Service')
    return false
  }

  return true
}

const handleSignUp = async () => {
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
        notificationStore.success('Account Created!', 'Welcome to KubeCloud! Please check your email to verify your account.')
        router.push('/sign-in')
      },
      'Creating your account...',
      'Account created successfully!'
    )
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Sign up failed'
    notificationStore.error('Sign Up Failed', message)
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
  background: var(--color-bg);
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
