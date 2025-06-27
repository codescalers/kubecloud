<template>
  <div class="dashboard-card card-enhanced">
    <div class="card-header">
      <div class="card-title-section">
        <div class="card-title-content">
          <h3 class="dashboard-card-title">Profile Settings</h3>
          <p class="card-subtitle">Manage your account information and preferences</p>
        </div>
      </div>
    </div>
    <v-form @submit.prevent="saveProfile" class="profile-form">
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="profile.firstName"
            label="First Name"
            variant="outlined"
            class="profile-field"
            :rules="[rules.required]"
            color="accent"
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
            color="accent"
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
            color="accent"
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
            color="accent"
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
            color="accent"
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
            color="accent"
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
            color="accent"
            bg-color="transparent"
            hide-details="auto"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" class="d-flex justify-end">
          <v-btn type="submit" variant="outlined" color="white" class="action-btn">Save Changes</v-btn>
        </v-col>
      </v-row>
    </v-form>
  </div>
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
.profile-form {
  width: 100%;
}

.profile-field {
  color: #CBD5E1;
  background: rgba(96, 165, 250, 0.1);
  border: 1px solid rgba(96, 165, 250, 0.2);
  border-radius: 0.75rem;
  margin-bottom: 1rem;
  transition: all 0.2s ease;
}

.profile-field:hover {
  border-color: rgba(96, 165, 250, 0.4);
  background: rgba(96, 165, 250, 0.15);
}

.profile-field:focus-within {
  border-color: rgba(96, 165, 250, 0.6);
}

/* Vuetify field overrides */
.profile-field :deep(.v-field) {
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

.profile-field :deep(.v-field__input) {
  color: #fff !important;
  font-size: 1rem;
}

.profile-field :deep(.v-field__outline) {
  display: none !important;
}

.profile-field :deep(.v-label) {
  color: #60a5fa !important;
  font-weight: 500;
}

.profile-field :deep(.v-field--focused .v-label) {
  color: #60a5fa !important;
}

.profile-field :deep(.v-field--variant-outlined .v-field__outline__start) {
  border-color: transparent !important;
}

.profile-field :deep(.v-field--variant-outlined .v-field__outline__end) {
  border-color: transparent !important;
}

.profile-field :deep(.v-field--variant-outlined .v-field__outline__notch) {
  border-color: transparent !important;
}

/* Button styling */
.action-btn {
  background: transparent !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  color: #fff !important;
  font-weight: 500;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1) !important;
  border-color: rgba(255, 255, 255, 0.6) !important;
}

@media (max-width: 960px) {
  .profile-form {
    margin-top: 1rem;
  }
}

@media (max-width: 600px) {
  .profile-field {
    margin-bottom: 0.75rem;
  }
}
</style>
