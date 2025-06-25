<template>
  <div class="dashboard-card profile-card">
    <h2 class="dashboard-card-title mb-6">Profile Settings</h2>
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
          <v-btn type="submit" class="action-btn">Save Changes</v-btn>
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
.dashboard-card.profile-card {
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  color: var(--color-text);
  padding: 1.25rem;
  min-width: 100%;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}
.profile-form {
  width: 100%;
}
.profile-field {
  color: var(--color-text);
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  margin-bottom: 1rem;
}
.action-btn {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}
.action-btn:hover {
  background: var(--color-accent-hover);
}
</style>
