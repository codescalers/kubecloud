<template>
  <v-form @submit.prevent="confirm($event.target as HTMLFormElement)">
    <v-card :style="{ padding: '0 !important' }">
      <v-card-title class="px-6 py-4">
        <div class="d-flex align-center justify-space-between">
          <h3 class="text-h5 font-weight-bold">
            Add Link
          </h3>
        </div>
      </v-card-title>

      <v-divider />

      <v-card-text>
        <v-row>
          <v-col cols="12">
            <v-text-field
              label="SSH Key Name"
              name="name"
              placeholder="Enter the ssh key name"
              variant="outlined"
              hide-details
              autofocus
            />
            <!-- :rules="[
                (v) => !!v || 'Key name is required',
                (v) => v.length >= 3 || 'Key name must be at least 3 characters',
                (v) => v.length <= 255 || 'Key name must be less than 255 characters',
              ]" -->
          </v-col>

          <v-col cols="12">
            <v-textarea
              label="Public SSH Key"
              name="public_key"
              placeholder="Enter the public ssh key"
              variant="outlined"
              hide-details
              no-resize
              rows="3"
            />
          </v-col>

          <v-col cols="12">
            <VBtn
              type="button"
              tabindex="-1"
              prepend-icon="mdi-key-plus"
              variant="plain"
              border
              color="primary"
              text="Generate SSH Key"
              @click="console.warn('generate ssh key button clicked')"
            />
          </v-col>
        </v-row>
      </v-card-text>
      <v-divider />

      <v-card-actions class="px-6 py-4 flex-row-reverse justify-start">
        <v-btn variant="text" color="primary" type="submit">
          Add
        </v-btn>
        <v-btn variant="plain" type="button" @click="$emit('cancel')">
          Cancel
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-form>
</template>

<script setup lang="ts">
import type { HandlersSSHKeyInput } from "~/generated/api"

const emit = defineEmits<{
  (e: "confirm", value: HandlersSSHKeyInput): void
  (e: "cancel"): void
}>()

function confirm(event: HTMLFormElement) {
  const form = new FormData(event)
  const name = form.get("name") as string
  const public_key = form.get("public_key") as string
  emit("confirm", { name, public_key })
}
</script>
