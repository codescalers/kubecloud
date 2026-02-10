<template>
  <div>
    <h1 class="text-h5 font-weight-bold">
      SSH Keys
    </h1>
    <p class="text-body-2 mt-1 text-accent">
      Manage your SSH keys for secure server access accross your infrastructure.
    </p>

    <div class="mt-8 d-flex justify-end">
      <VBtn
        variant="tonal"
        color="primary"
        prepend-icon="mdi-plus"
        text="Add SSH Key"
        :loading="isLoading"
        @click="onAddSSHKey()"
      />
    </div>

    <VCard class="mt-4" :style="{ padding: '0 !important' }">
      <VCardText v-if="keys.length === 0" class="text-body-2 text-center text-accent py-8">
        No SSH keys found.
      </VCardText>

      <VCardText
        v-for="(key, index) in keys"
        :key="key.id"
        :class="{ 'border-t': index !== 0 }"
      >
        <SSHKey :public-key="key" @delete="keys = keys.filter((_, i) => i !== index)" />
      </VCardText>
    </VCard>

    <v-dialog :model-value="isRevealed" max-width="500" scrollable @update:model-value="cancel()">
      <AddSSHKeyDialogCard @confirm="confirm($event)" />
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { HandlersSSHKeyInput } from "~/generated/api"

const api = useApi()

const { state: keys } = useAsyncState(async () => {
  const { data } = await api.users.listSshKeys()
  return data.data ?? []
}, [], { immediate: $meta.client })

const toast = useToast()
const { execute: addSSHKey, isLoading } = useAsyncState(async (value: HandlersSSHKeyInput) => {
  const { data } = await api.users.addSshKey(value)
  toast.success({ message: data.message })
  if (data.data) {
    keys.value = [...keys.value, data.data]
  }
}, null, { immediate: false })

const { isRevealed, reveal, cancel, confirm } = useDialog<undefined, HandlersSSHKeyInput>()
async function onAddSSHKey() {
  const { data, isCanceled } = await reveal()
  if (!isCanceled && data) {
    addSSHKey(undefined, data!)
  }
}
</script>
