<template>
  <v-dialog v-model="dialog" max-width="800px" scrollable>
    <v-card>
      <v-card-title class="d-flex align-center justify-space-between">
        <span class="text-h5">Terms and Conditions</span>
        <v-btn icon variant="text" @click="dialog = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>
      
      <v-divider></v-divider>
      
      <v-card-text class="pa-6" style="height: 500px;">
        <div class="terms-content">
          <h3>1. Acceptance of Terms</h3>
          <p>
            By accessing and using Mycelium Cloud services, you accept and agree to be bound by the terms and provision of this agreement.
          </p>

          <h3>2. Service Description</h3>
          <p>
            Mycelium Cloud provides cloud computing services including but not limited to virtual machines, clusters, and storage solutions.
          </p>

          <h3>3. User Responsibilities</h3>
          <p>
            Users are responsible for:
          </p>
          <ul>
            <li>Maintaining the confidentiality of their account credentials</li>
            <li>All activities that occur under their account</li>
            <li>Ensuring compliance with applicable laws and regulations</li>
            <li>Proper use of allocated resources</li>
          </ul>

          <h3>4. Service Availability</h3>
          <p>
            We strive to maintain high service availability but do not guarantee uninterrupted service. Scheduled maintenance will be communicated in advance.
          </p>

          <h3>5. Data Privacy and Security</h3>
          <p>
            We implement industry-standard security measures to protect your data. However, users are responsible for implementing appropriate security measures for their applications and data.
          </p>

          <h3>6. Billing and Payment</h3>
          <p>
            Services are billed according to usage and selected plans. Payment is due according to the billing cycle selected during registration.
          </p>

          <h3>7. Limitation of Liability</h3>
          <p>
            Mycelium Cloud shall not be liable for any indirect, incidental, special, consequential, or punitive damages resulting from your use of the service.
          </p>

          <h3>8. Termination</h3>
          <p>
            Either party may terminate this agreement at any time. Upon termination, access to services will be discontinued and data may be deleted according to our data retention policy.
          </p>

          <h3>9. Changes to Terms</h3>
          <p>
            We reserve the right to modify these terms at any time. Users will be notified of significant changes via email or service notifications.
          </p>

          <h3>10. Contact Information</h3>
          <p>
            For questions about these terms, please contact us at support@myceliumcloud.com
          </p>

          <p class="mt-4 text-caption text-medium-emphasis">
            Last updated: {{ new Date().toLocaleDateString() }}
          </p>
        </div>
      </v-card-text>
      
      <v-divider></v-divider>
      
      <v-card-actions class="pa-4">
        <v-spacer></v-spacer>
        <v-btn color="primary" variant="outlined" @click="dialog = false">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Props {
  modelValue: boolean
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const dialog = ref(props.modelValue)

watch(() => props.modelValue, (newValue) => {
  dialog.value = newValue
})

watch(dialog, (newValue) => {
  emit('update:modelValue', newValue)
})
</script>

<style scoped>
.terms-content h3 {
  color: rgb(var(--v-theme-primary));
  margin-top: 1.5rem;
  margin-bottom: 0.5rem;
  font-weight: 600;
}

.terms-content h3:first-child {
  margin-top: 0;
}

.terms-content p {
  margin-bottom: 1rem;
  line-height: 1.6;
}

.terms-content ul {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
}

.terms-content li {
  margin-bottom: 0.25rem;
  line-height: 1.5;
}
</style>
