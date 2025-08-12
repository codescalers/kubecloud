<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content">
      <div class="auth-header">
        <h1 class="auth-title">Verify Your Email</h1>
        <p class="auth-subtitle">Enter the verification code sent to your email</p>
      </div>
      <v-form @submit.prevent="handleVerify" class="auth-form">
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
          v-model="form.code"
          label="Verification Code"
          type="text"
          prepend-inner-icon="mdi-shield-key"
          variant="outlined"
          class="auth-field"
          :error-messages="errors.code"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :loading="loading"
          :disabled="resending"
        >
          <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
          Verify
        </v-btn>
      </v-form>
      <div class="auth-footer">
        <span class="auth-footer-text">Didn't receive a code?</span>
        <v-btn variant="outlined" color="white" :disabled="loading" :loading="resending" @click="resendCode">Resend Code</v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authService } from '../utils/authService'

const route = useRoute()
const router = useRouter()

const form = reactive({
  email: route.query.email ? String(route.query.email) : '',
  code: ''
})

const errors = reactive({
  email: '',
  code: ''
})
const loading = ref(false)
const resending = ref(false)

const clearErrors = () => {
  errors.email = ''
  errors.code = ''
}

const handleVerify = async () => {
  clearErrors()
  if (!form.email) {
    errors.email = 'Email is required'
    return
  }
  if (!form.code) {
    errors.code = 'Verification code is required'
    return
  }
  try {
    loading.value = true
    await authService.verifyCode({ email: form.email, code: Number(form.code) })
    loading.value = false
    router.push('/sign-in')
  } catch (error) {
    loading.value = false
    console.error(error)
  }
}

const resendCode = async () => {
  resending.value = true
  if (!form.email) {
    errors.email = 'Email is required to resend code'
    return
  }
  try {
    await authService.register({
      name: 'User',
      email: form.email,
      password: 'temporary',
      confirm_password: 'temporary'
    })
  } catch (error) {
    console.error(error)
  } finally {
    resending.value = false
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
</style>
