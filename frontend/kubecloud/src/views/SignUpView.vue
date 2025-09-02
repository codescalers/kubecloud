<template>
  <div class="auth-view">
    <LoadingComponent v-if="loading" fullPage message="Creating account..." />
    <div class="auth-content">
      <div class="auth-header">
        <h1 class="auth-title">Create Account</h1>
        <p class="auth-subtitle">Join KubeCloud and start your journey</p>
      </div>
      <v-form @submit.prevent="handleSignUp" class="auth-form" v-model="isFormValid">
        <v-text-field
          v-model="form.name"
          label="Username"
          prepend-inner-icon="mdi-account"
          variant="outlined"
          class="auth-field"
          :rules="[RULES.username]"
          required
        />
        <v-text-field
          v-model="form.email"
          label="Email Address"
          type="email"
          prepend-inner-icon="mdi-email"
          variant="outlined"
          class="auth-field"
          :rules="[RULES.email]"
          required
        />
        <v-text-field
          v-model="form.password"
          label="Password"
          :type="showPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append-inner="showPassword = !showPassword"
          variant="outlined"
          class="auth-field"
          :rules="[RULES.password]"
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
          :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append-inner="showConfirmPassword = !showConfirmPassword"
          variant="outlined"
          class="auth-field"
          :rules="[RULES.confirmPassword(form.confirmPassword, form.password)]"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :disabled="loading || !isFormValid"
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
      <router-link
        to="/"
        class="d-block text-white back-home-link"
      >
        Back to Home
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { RULES } from '../utils/validation'
import LoadingComponent from '../components/LoadingComponent.vue'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const isFormValid = ref(false)
const form = reactive({
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const showPassword = ref(false)
const showConfirmPassword = ref(false)


const handleSignUp = async () => {
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

.back-home-link {
  margin-top: 1rem;
  text-decoration: none;
}
.back-home-link:hover {
  text-decoration: underline;
}
</style>
