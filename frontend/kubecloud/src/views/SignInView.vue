<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content">
      <div class="auth-header">
        <h1 class="auth-title">Welcome Back!</h1>
        <p class="auth-subtitle">Sign in to your KubeCloud account</p>
      </div>
      <v-form @submit.prevent="handleSignIn" class="auth-form" ref="formRef" v-model="isFormValid">
        <v-text-field
          v-model="form.email"
          label="Email Address"
          type="email"
          prepend-inner-icon="mdi-email"
          variant="outlined"
          class="auth-field"
          :disabled="loading"
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
          :rules="[RULES.password]"
          variant="outlined"
          class="auth-field"
          :disabled="loading"
          required
        />
        <div class="auth-options-vertical">
          <v-btn
            variant="text"
            size="small"
            class="kubecloud-hover-blue pa-0"
            :disabled="loading"
            @click="router.push('/forgot-password')"
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
          :disabled="loading || !isFormValid"
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

            <router-link
              to="/"
              class="text-white back-home-link"
            >
              Back to Home
            </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { RULES } from '../utils/validation'

const router = useRouter()
const userStore = useUserStore()

const form = reactive({
  email: '',
  password: '',
})


const showPassword = ref(false)
const loading = ref(false)
const isFormValid = ref(false)

const handleSignIn = async () => {
  loading.value = true
  try {
    await userStore.login(form.email, form.password)
    await nextTick()
    try {
      await router.replace('/')
    } catch (routerError) {
      window.location.href = '/'
    }
  } catch (error) {
    console.error('Sign in error:', error)
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

.back-home-link {
  display: block;
  margin-top: 1rem;
  text-decoration: none;
}
.back-home-link:hover {
  text-decoration: underline;
}
</style>
