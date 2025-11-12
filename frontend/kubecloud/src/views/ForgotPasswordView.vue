<template>
  <div class="auth-view">
    <div class="auth-background"></div>
    <div class="auth-content">
      <div class="auth-header">
        <h1 class="auth-title">Forgot Password</h1>
        <p class="auth-subtitle">{{ step === 1 ? 'Enter your email to receive a reset code.' : 'Enter the verification code sent to your email.' }}</p>
      </div>
      <v-form v-if="step === 1" @submit.prevent="handleRequestCode" class="auth-form" v-model="isEmailFormValid">
        <v-text-field
          v-model="email"
          label="Email Address"
          type="email"
          prepend-inner-icon="mdi-email"
          variant="outlined"
          class="auth-field"
          :disabled="loading"
          :rules="[RULES.email]"
          required
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :loading="loading"
          :disabled="loading || !isEmailFormValid"
        >
          <v-icon icon="mdi-email-send" class="mr-2"></v-icon>
          {{ loading ? 'Sending...' : 'Send Reset Code' }}
        </v-btn>
      </v-form>
      <v-form v-else @submit.prevent="handleVerifyCode" class="auth-form" v-model="isCodeFormValid">
        <v-text-field
          v-model="code"
          label="Verification Code"
          type="number"
          prepend-inner-icon="mdi-numeric"
          variant="outlined"
          class="auth-field"
          :disabled="loading"
          :rules="[RULES.verificationCode]"
          required
          placeholder="Enter 4-6 digit code"
          maxlength="6"
        />
        <v-btn
          type="submit"
          color="white"
          block
          size="large"
          variant="outlined"
          :loading="loading"
          :disabled="loading || !isCodeFormValid"
        >
          <v-icon icon="mdi-check" class="mr-2"></v-icon>
          {{ loading ? 'Verifying...' : 'Verify Code' }}
        </v-btn>
      </v-form>
        <router-link
          to="/sign-in"
          class="text-white back-home-link justify-center"
        >
          <v-icon icon="mdi-arrow-left" size="16" class="mr-1 text-center"></v-icon>
          Back to Sign In
        </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authService } from '../utils/authService'
import { RULES } from '../utils/validation'

const router = useRouter()
const step = ref(1)
const email = ref('')
const code = ref('')
const loading = ref(false)
const isCodeFormValid = ref(false)
const isEmailFormValid = ref(false)

const handleRequestCode = async () => {
  loading.value = true
  try {
    await authService.forgotPassword({ email: email.value.trim() })
    step.value = 2
  } catch (err: any) {
    console.error('Failed to send reset code:', err)
  } finally {
    loading.value = false
  }
}

const handleVerifyCode = async () => {
  if (!code.value.trim() || RULES.verificationCode(code.value) !== true) return
  loading.value = true
  try {
    const tokens = await authService.verifyForgotPasswordCode({
      email: email.value.trim(),
      code: Number(code.value.trim())
    })

    authService.storeTempTokens(tokens.access_token, tokens.refresh_token)
    localStorage.setItem('password_reset_session', 'true')

    router.push({
      path: '/reset-password',
      query: { email: email.value.trim() }
    })
  } catch (err: any) {
    console.error('Invalid verification code:', err)
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
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  letter-spacing: -0.5px;
  line-height: 1.1;
}
.auth-subtitle {
  font-size: var(--font-size-md);
  color: #fff;
  opacity: 0.92;
  font-weight: var(--font-weight-normal);
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

.back-home-link {
  display: flex;
  align-items: center;
  margin-top: 1rem;
  text-decoration: none;
  transition: all 0.2s ease;
}
.back-home-link:hover {
  text-decoration: underline;
  opacity: 0.8;
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
