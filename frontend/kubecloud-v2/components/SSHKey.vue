<template>
  <div class="d-flex ga-4">
    <VAvatar
      size="x-large"
      rounded="lg"
      variant="tonal"
      icon="mdi-key-outline"
      color="primary"
    />

    <div class="flex-grow-1">
      <p class="text-subtitle-1 font-weight-bold" v-text="publicKey.name" />

      <div class="d-flex align-center ga-2 mt-2 w-100">
        <p
          class="text-subtitle-2 border d-inline-block px-3 py-1 rounded-lg w-100"
          :style="{
            backgroundColor: 'rgb(var(--v-bg-2))',
            maxWidth: '600px',
            width: '100%',
          }"
        >
          <span class="text-accent" v-text="formattedPublicKey" />
        </p>

        <VBtn
          icon
          size="x-small"
          variant="plain"
          @click="console.log('copied')"
        >
          <VIcon icon="mdi-content-copy" size="small" color="accent" />
        </VBtn>
      </div>

      <div class="d-flex align-center ga-1 mt-1">
        <VIcon icon="mdi-calendar-outline" size="x-small" color="accent" />
        <span class="text-subtitle-2 text-accent">Added on {{ addedAt }}</span>
      </div>
    </div>

    <div class="align-self-center">
      <VBtn
        variant="text"
        prepend-icon="mdi-trash-can-outline"
        color="error"
        size="small"
        text="Remove"
        :loading="isLoading"
        @click="deleteKey()"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ModelsSSHKey } from "~/generated/api"

const props = defineProps<{ publicKey: ModelsSSHKey }>()
const emit = defineEmits<{ (e: "delete"): void }>()

const formattedPublicKey = computed(() => {
  const key = props.publicKey.public_key
  return `${key.slice(0, 25)} ... ${key.slice(-25)}`
})

const addedAt = useDateFormat(() => props.publicKey.created_at, "MMM d, YYYY")

const api = useApi()
const toast = useToast()
const { isLoading, execute: deleteKey } = useAsyncState(async () => {
  const { data } = await api.users.deleteSshKey((props.publicKey as any).ID)
  toast.success({ message: data.message })
  emit("delete")
}, null, { immediate: false })
</script>
