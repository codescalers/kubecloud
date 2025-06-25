<template>
  <div class="auth-container">
    <div class="auth-background">
      <div class="auth-particles"></div>
    </div>

    <v-card class="auth-card card-enhanced">
      <div class="auth-header">
        <h1 class="auth-title kubecloud-gradient">Create Account</h1>
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
          color="secondary"
          block
          size="large"
          class="auth-submit-btn btn-enhanced kubecloud-glow-orange"
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
          variant="text"
          color="primary"
          to="/sign-in"
          class="auth-link kubecloud-hover-blue"
          :disabled="loading"
        >
          Sign In
        </v-btn>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
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

.name-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1rem;
}

.auth-field {
  margin-bottom: 1.25rem;
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

.auth-options {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.terms-checkbox {
  margin-bottom: 0;
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

  .name-fields {
    grid-template-columns: 1fr;
  }

  .auth-options {
    gap: 0.5rem;
  }
}
</style>
