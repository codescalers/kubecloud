<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content">
      <div class="auth-header">
        <h1 class="auth-title">Forgot Password</h1>
        <p class="auth-subtitle">{{ step === 1 ? 'Enter your email to receive a reset code.' : 'Enter the verification code sent to your email.' }}</p>
      </div>
      <v-form v-if="step === 1" @submit.prevent="handleRequestCode" class="auth-form">
        <v-text-field
          v-model="email"
          label="Email Address"
          type="email"
          prepend-inner-icon="mdi-email"
          variant="outlined"
          class="auth-field"
          :error-messages="error"
          :disabled="loading"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :loading="loading"
          :disabled="loading || !isEmailValid"
        >
          <v-icon icon="mdi-email-send" class="mr-2"></v-icon>
          {{ loading ? 'Sending...' : 'Send Reset Code' }}
        </v-btn>
      </v-form>
      <v-form v-else @submit.prevent="handleVerifyCode" class="auth-form">
        <v-text-field
          v-model="code"
          label="Verification Code"
          type="number"
          prepend-inner-icon="mdi-numeric"
          variant="outlined"
          class="auth-field"
          :error-messages="error"
          :disabled="loading"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :loading="loading"
          :disabled="loading || !isCodeValid"
        >
          <v-icon icon="mdi-check" class="mr-2"></v-icon>
          {{ loading ? 'Verifying...' : 'Verify Code' }}
        </v-btn>
      </v-form>
      <div class="auth-footer">
        <v-btn variant="text" color="white" to="/sign-in">Back to Sign In</v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authService } from '@/utils/authService'

const router = useRouter()
const step = ref(1)
const email = ref('')
const code = ref('')
const loading = ref(false)
const error = ref('')

// Form validation
const isEmailValid = computed(() => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return email.value.trim() !== '' && emailRegex.test(email.value)
})

const isCodeValid = computed(() => {
  return code.value.trim() !== '' && code.value.length >= 4
})

const handleRequestCode = async () => {
  if (!isEmailValid.value) return

  error.value = ''
  loading.value = true
  try {
    await authService.forgotPassword({ email: email.value })
    step.value = 2
  } catch (err: any) {
    error.value = err?.message || 'Failed to send reset code'
  } finally {
    loading.value = false
  }
}

const handleVerifyCode = async () => {
  if (!isCodeValid.value) return

  error.value = ''
  loading.value = true
  try {
    // Verify the code and get tokens
    const tokens = await authService.verifyForgotPasswordCode({ email: email.value, code: Number(code.value) })

    // Store tokens temporarily for password reset only (separate from main auth)
    authService.storeTempTokens(tokens.access_token, tokens.refresh_token)

    // Mark this as a password reset session
    localStorage.setItem('password_reset_session', 'true')

    // Redirect to reset password page with email
    router.push({
      path: '/reset-password',
      query: {
        email: email.value
      }
    })
  } catch (err: any) {
    error.value = err?.message || 'Invalid code'
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
  position: relative;
  z-index: 1;
  background: rgba(10, 25, 47, 0.95);
  border-radius: 2rem;
  box-shadow: 0 8px 32px 0 rgba(16, 42, 67, 0.25);
  padding: 3rem 2.5rem 2.5rem 2.5rem;
  max-width: 400px;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  animation: fadeInUp 0.6s ease-out;
}
.auth-header {
  text-align: center;
}
.auth-title {
  font-size: 2.2rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 0.5rem;
}
.auth-subtitle {
  color: #fff;
  font-size: 1.1rem;
}
.auth-form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.auth-field {
  width: 100%;
}
.auth-footer {
  margin-top: 2rem;
  text-align: center;
}
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
