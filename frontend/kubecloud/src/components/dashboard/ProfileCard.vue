<template>
  <div class="dashboard-card profile-card compact">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Profile</h3>
      <p class="dashboard-card-subtitle">Your account information</p>
    </div>
    <div class="dashboard-card-content">
      <div v-if="user" class="profile-form">
        <div class="profile-row">
          <div class="profile-col">
            <label class="profile-label">Username</label>
            <v-text-field :model-value="user.username" variant="outlined" class="profile-field compact" color="accent" bg-color="transparent" hide-details="auto" disabled density="compact" />
          </div>
          <div class="profile-col">
            <label class="profile-label">Email Address</label>
            <v-text-field :model-value="user.email" variant="outlined" type="email" class="profile-field compact" color="accent" bg-color="transparent" hide-details="auto" disabled density="compact" />
          </div>
        </div>
        <div class="profile-row">
          <div class="profile-col">
            <label class="profile-label">Balance</label>
            <v-text-field :model-value="`$${userStore.netBalance.toFixed(2)}${userStore.pendingBalance > 0 ? ` (+$${userStore.pendingBalance.toFixed(2)} pending)` : ''}`" variant="outlined" class="profile-field compact" color="accent" bg-color="transparent" hide-details="auto" disabled density="compact" />
          </div>
          <div class="profile-col">
            <label class="profile-label">Verified</label>
            <v-text-field :model-value="user.verified ? 'Yes' : 'No'" variant="outlined" class="profile-field compact" color="accent" bg-color="transparent" hide-details="auto" disabled density="compact" />
          </div>
        </div>
        <v-divider class="my-4"></v-divider>
        <div class="change-password-section">
          <h4 class="section-title">Change Password</h4>
            <v-form
            @submit.prevent="changePassword"
            ref="passwordForm"
            v-model="isFormValid"
            >
            <div class="profile-row">
              <div class="profile-col">
              <v-text-field
                v-model="passwordFormData.password"
                :type="showPassword ? 'text' : 'password'"
                label="New Password"
                variant="outlined"
                :rules="passwordRules"
                required
                class="password-field compact"
                density="compact"
                :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                @click:append-inner="() => { showPassword = !showPassword }"
                @input="triggerFormValidation"
              />
              </div>
              <div class="profile-col">
              <v-text-field
                v-model="passwordFormData.confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                label="Confirm New Password"
                variant="outlined"
                :rules="confirmPasswordRules"
                required
                class="password-field compact"
                density="compact"
                :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
                @click:append-inner="() => { showConfirmPassword = !showConfirmPassword }"
                @input="triggerFormValidation"
              />
              </div>
            </div>
            <v-btn
              type="submit"
              color="primary"
              variant="outlined"
              :loading="loading"
              :disabled="loading || !isFormValid"
              prepend-icon="mdi-lock-reset"
              class="action-btn compact"
            >
              Change Password
            </v-btn>
          </v-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '../../stores/user'
import { authService } from '../../utils/authService'

const { user } = storeToRefs(useUserStore())
const userStore = useUserStore()

// Change password form data
const passwordFormData = ref({
  email: '',
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const isFormValid = ref(false)

// Show/hide password toggles
const showPassword = ref(false)
const showConfirmPassword = ref(false)

// Form validation rules
const passwordRules = [
  (v: string) => !!v || 'Password is required',
  (v: string) => v.length >= 8 || 'Password must be at least 8 characters',
  (v: string) => /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(v) || 'Password must contain uppercase, lowercase, and number'
]

const confirmPasswordRules = [
  (v: string) => !!v || 'Please confirm your password',
  (v: string) => v === passwordFormData.value.password || 'Passwords do not match'
]

function triggerFormValidation() {
  if (passwordForm.value && passwordForm.value.validate) {
    passwordForm.value.validate()
  }
}

watch(
  () => [passwordFormData.value.password, passwordFormData.value.confirmPassword],
  triggerFormValidation
)

const passwordForm = ref()

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleString()
}

async function changePassword() {
  if (!user.value) {
    return
  }

  loading.value = true

  try {
    await authService.changePassword({
      email: user.value.email,
      password: passwordFormData.value.password,
      confirm_password: passwordFormData.value.confirmPassword
    })
    // Clear form
    passwordFormData.value.password = ''
    passwordFormData.value.confirmPassword = ''
    showPassword.value = false
    showConfirmPassword.value = false
    // Reset validation state
    if (passwordForm.value && passwordForm.value.resetValidation) {
      passwordForm.value.resetValidation()
    }
  } catch (err: any) {
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.dashboard-card.profile-card.compact {
  padding: 1.5rem;
}

.dashboard-card-header {
  margin-bottom: 1rem;
}

.dashboard-card-title {
  font-size: 1.25rem;
  margin-bottom: 0.25rem;
}

.dashboard-card-subtitle {
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
}

.profile-form {
  width: 100%;
}

.profile-field.compact {
  color: #94a3b8;
  background: #232946;
  border: none;
  border-radius: 0.5rem;
  margin-bottom: 0.75rem;
  transition: none;
  opacity: 1;
}

.profile-field.compact:focus-within,
.profile-field.compact:hover {
  border: none;
  box-shadow: none;
}

.profile-field.compact[readonly],
.profile-field.compact :deep(input[readonly]) {
  background: rgba(96, 165, 250, 0.08) !important;
  border-color: rgba(96, 165, 250, 0.12) !important;
  cursor: default !important;
}

.profile-field.compact :deep(.v-field) {
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
  min-height: 40px !important;
}

.profile-field.compact :deep(.v-field__input) {
  color: #fff !important;
  font-size: 0.9rem;
  padding: 8px 12px !important;
}

.profile-field.compact :deep(.v-field__outline) {
  display: none !important;
}

.profile-field.compact :deep(.v-field--focused .v-label) {
  color: #60a5fa !important;
}

.profile-field.compact :deep(.v-field--variant-outlined .v-field__outline__start) {
  border-color: transparent !important;
}

.profile-field.compact :deep(.v-field--variant-outlined .v-field__outline__end) {
  border-color: transparent !important;
}

.profile-field.compact :deep(.v-field--variant-outlined .v-field__outline__notch) {
  border-color: transparent !important;
}

.profile-label {
  display: block;
  margin-bottom: 0.25rem;
  color: var(--color-text, #fff);
  font-weight: 500;
  font-size: 0.85rem;
}

.change-password-section {
  margin-top: 1.5rem;
}

.section-title {
  color: var(--color-text, #fff);
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.password-field.compact :deep(.v-field) {
  background: rgba(96, 165, 250, 0.08) !important;
  border: 1px solid rgba(96, 165, 250, 0.12) !important;
  border-radius: 0.5rem !important;
  min-height: 40px !important;
}

.password-field.compact :deep(.v-field__input) {
  color: #fff !important;
  padding: 8px 12px !important;
}

.password-field.compact :deep(.v-label) {
  color: #60a5fa !important;
  font-weight: 500;
  font-size: 0.9rem;
}

.profile-row {
  display: flex;
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
  gap: 0.75rem;
}

.profile-col {
  flex: 1;
  min-width: 200px;
}

.action-btn.compact {
  height: 40px;
  font-size: 0.9rem;
}

@media (max-width: 768px) {
  .dashboard-card.profile-card.compact {
    padding: 1rem;
  }
  .profile-row {
    flex-direction: column;
    gap: 0.5rem;
  }
  .profile-col {
    min-width: auto;
  }
}
</style>
