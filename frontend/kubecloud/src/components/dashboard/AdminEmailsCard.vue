<template>
  <v-card class="pa-6" elevation="2">
    <v-card-title class="text-h4 pa-0 mb-2">
      Emails
    </v-card-title>
    <v-card-subtitle class="pa-0 mb-6">
      Send emails to all platform users
    </v-card-subtitle>

    <v-alert type="warning" variant="tonal" class="my-4">
      This will send an email to all registered users on the platform.
    </v-alert>
    <v-card-text class="pa-0">
      <v-form ref="emailForm" @submit.prevent="sendEmail">
        <v-text-field v-model="title" label="Email Subject" placeholder="Email Subject" variant="outlined" class="mb-4"
          :rules="[v => !!v || 'Subject is required']"></v-text-field>



        <v-textarea v-model="message" label="Email Content" placeholder="Compose your email message here..."
          variant="outlined" rows="6" class="mb-4" :rules="[v => !!v || 'Content is required']"></v-textarea>

        <v-file-input v-model="attachments" label="Attachments (optional)" variant="outlined" multiple show-size counter
          class="mb-4" accept=".pdf,.doc,.docx,.txt,.jpg,.jpeg,.png,.gif,.zip" :rules="attachmentRules">
          <template v-slot:selection="{ fileNames }">
            <template v-for="(fileName, index) in fileNames" :key="fileName">
              <v-chip v-if="index < 2" color="primary" size="small" class="me-2">
                {{ fileName }}
              </v-chip>
              <span v-else-if="index === 2" class="text-overline grey--text text--darken-3 mx-2">
                +{{ fileNames.length - 2 }} File(s)
              </span>
            </template>
          </template>
        </v-file-input>

        <v-btn type="submit" color="primary" :loading="sending" :disabled="!title || !message || sending"
          :size="$vuetify.display.xs ? 'default' : 'large'" :block="$vuetify.display.xs" class="mt-4">
          <span class="d-none d-sm-inline">Send to All Users</span>
          <span class="d-sm-none">Send Email</span>
        </v-btn>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { adminService } from '@/utils/adminService'
import { useNotificationStore } from '@/stores/notifications'

// Form state
const emailForm = ref()
const title = ref('')
const message = ref('')
const attachments = ref<File[]>([])
const sending = ref(false)

// File validation rules
const attachmentRules = [
  (files: File[]) => {
    if (!files || files.length === 0) return true
    const maxSize = 10 * 1024 * 1024 // 10MB
    const oversized = files.find(file => file.size > maxSize)
    if (oversized) return `File "${oversized.name}" is too large. Maximum size is 10MB.`
    return true
  },
  (files: File[]) => {
    if (!files || files.length === 0) return true
    if (files.length > 5) return 'Maximum 5 files allowed'
    return true
  }
]

// Methods
async function sendEmail() {
  if (!title.value || !message.value) return

  sending.value = true
  try {
    const formData = new FormData()
    formData.append('subject', title.value)
    formData.append('body', message.value)

    if (attachments.value && attachments.value.length > 0) {
      attachments.value.forEach((file) => {
        formData.append(`attachments`, file)
      })
    }

    const response = await adminService.sendSystemEmail(formData)
    if (response.failed_emails_count > 0) {
      useNotificationStore().warning('Not all users received the email', `Failed to send email to ${response.failed_emails_count} users`)
    }

    emailForm.value?.reset()
    title.value = ''
    message.value = ''
    attachments.value = []
  } catch (error) {
    console.error(error)
  } finally {
    sending.value = false
  }
}
</script>
