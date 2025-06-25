<template>
  <v-card class="pa-8 profile-card card-enhanced">
    <h2 class="display-3 mb-6 neon-glow-pink">Profile Settings</h2>
    <v-form @submit.prevent="saveProfile" class="profile-form">
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.firstName"
            label="First Name"
            variant="outlined"
            class="profile-field"
            :rules="[rules.required]"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.lastName"
            label="Last Name"
            variant="outlined"
            class="profile-field"
            :rules="[rules.required]"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.email"
            label="Email Address"
            variant="outlined"
            type="email"
            class="profile-field"
            :rules="[rules.required, rules.email]"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.phone"
            label="Phone Number"
            variant="outlined"
            class="profile-field"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12">
          <v-textarea
            v-model="profile.bio"
            label="Bio"
            variant="outlined"
            rows="3"
            class="profile-field"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.company"
            label="Company"
            variant="outlined"
            class="profile-field"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.website"
            label="Website"
            variant="outlined"
            class="profile-field"
            color="primary"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" class="d-flex justify-end">
          <v-btn
            type="submit"
            variant="elevated"
            color="primary"
            size="large"
            class="save-btn neon-glow-pink"
            :loading="loading"
          >
            <v-icon icon="mdi-content-save" size="24" class="mr-2"></v-icon>
            Save Changes
          </v-btn>
        </v-col>
      </v-row>
    </v-form>
  </v-card>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const loading = ref(false)

const profile = reactive({
  firstName: 'John',
  lastName: 'Doe',
  email: 'john.doe@example.com',
  phone: '+1 (555) 123-4567',
  bio: 'Full-stack developer passionate about cloud computing and decentralized technologies.',
  company: 'TechCorp Inc.',
  website: 'https://johndoe.dev'
})

const rules = {
  required: (value: string) => !!value || 'This field is required.',
  email: (value: string) => {
    const pattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return pattern.test(value) || 'Please enter a valid email address.'
  }
}

function saveProfile() {
  loading.value = true
  // Simulate API call
  setTimeout(() => {
    loading.value = false
    // Show success message
  }, 2000)
}
</script>

<style scoped>
.profile-card {
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 1px solid var(--glass-border);
  border-radius: 24px;
  box-shadow: var(--shadow-md), var(--shadow-ambient);
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
  min-width: 100%;
}

.profile-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--kubecloud-orange-subtle), transparent);
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.profile-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl), var(--shadow-ambient);
  border-color: var(--kubecloud-orange-subtle);
}

.profile-card:hover::before {
  opacity: 1;
}

@keyframes shimmer {
  0%, 100% { transform: translateX(-100%); }
  50% { transform: translateX(100%); }
}

.profile-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.profile-field {
  background: rgba(234, 88, 12, 0.05) !important;
  border-radius: 12px !important;
  transition: all var(--transition-normal);
}

.profile-field:hover {
  background: rgba(234, 88, 12, 0.1) !important;
  box-shadow: var(--shadow-kubecloud-orange);
}

.profile-field :deep(.v-field) {
  background: transparent !important;
  border: 1px solid rgba(234, 88, 12, 0.3) !important;
  border-radius: 12px !important;
  transition: all var(--transition-normal);
}

.profile-field :deep(.v-field:hover) {
  border: 1px solid var(--kubecloud-orange) !important;
  box-shadow: var(--shadow-kubecloud-orange);
}

.profile-field :deep(.v-field--focused) {
  border: 2px solid var(--kubecloud-orange) !important;
  box-shadow: var(--shadow-hover-orange);
}

.profile-field :deep(.v-field__input) {
  color: #fff !important;
  font-size: 1rem !important;
  font-weight: 500 !important;
  padding: 1rem !important;
}

.profile-field :deep(.v-field__label) {
  color: var(--kubecloud-orange) !important;
  font-weight: 600 !important;
  font-size: 1rem !important;
  opacity: 0.9 !important;
}

.profile-field :deep(.v-field--focused .v-field__label) {
  color: var(--kubecloud-orange) !important;
  text-shadow: 0 0 8px var(--kubecloud-orange);
}

.save-btn {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-weight: 600;
  font-size: 1.125rem;
  padding: 1rem 2rem;
  border-radius: 12px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: linear-gradient(135deg, var(--kubecloud-orange), var(--kubecloud-orange-dark)) !important;
  box-shadow: var(--shadow-kubecloud-orange);
  transition: all var(--transition-normal);
  color: #fff !important;
  min-width: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 56px;
}

.save-btn:hover {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

.save-btn:active {
  transform: translateY(0);
}

@media (max-width: 960px) {
  .profile-card {
    padding: 2rem !important;
  }
  .profile-field :deep(.v-field__input) {
    padding: 0.875rem !important;
    font-size: 0.9rem !important;
  }
  .profile-field :deep(.v-field__label) {
    font-size: 0.9rem !important;
  }
  .save-btn {
    padding: 0.875rem 1.5rem;
    font-size: 1rem;
    min-width: 180px;
  }
}

@media (max-width: 600px) {
  .profile-card {
    padding: 1.5rem !important;
  }
  .profile-field :deep(.v-field__input) {
    padding: 0.75rem !important;
    font-size: 0.875rem !important;
  }
  .profile-field :deep(.v-field__label) {
    font-size: 0.875rem !important;
  }
  .save-btn {
    padding: 0.75rem 1rem;
    font-size: 0.9rem;
    min-width: 160px;
    width: 100%;
  }
}
</style>
