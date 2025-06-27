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
  margin-bottom: 0;
}
.auth-form {
  width: 100%;
}
.name-fields {
  display: flex;
  gap: 0.5rem;
}
@media (max-width: 600px) {
  .name-fields {
    flex-direction: column;
    gap: 0.3rem;
  }
}
.auth-options {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
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
