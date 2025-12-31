<template>
  <tr class="text-no-wrap py-2">
    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ user.id }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ user.username }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ user.email }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50"> ${{ Math.round(user.balance! * 100) / 100 }} </span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ createdAt }}</span>
    </td>

    <td>
      <div class="d-flex align-center justify-end ga-4">
        <v-btn
          variant="text"
          class="border"
          prepend-icon="mdi-cash-plus"
          size="small"
          text="Credit Balance"
          :disabled="!user.verified"
          :loading="isCreditLoading"
          @click="onCredit()"
        />

        <v-btn
          variant="text"
          class="border"
          color="warning"
          prepend-icon="mdi-water-remove"
          size="small"
          text="Drain"
          :loading="isDrainLoading"
          @click="onDrain()"
        />

        <v-btn
          variant="text"
          class="border"
          prepend-icon="mdi-trash-can-outline"
          color="error"
          size="small"
          text="Remove"
          :loading="isRemoveLoading"
          @click="onRemove()"
        />
      </div>
    </td>
  </tr>
</template>

<script setup lang="ts">
import type { HandlersCreditRequestInput, ServicesUserWithUSDBalance } from "../generated/api"

const props = defineProps<{ user: ServicesUserWithUSDBalance }>()
const emit = defineEmits<{ (e: "remove"): void }>()

const createdAt = useDateFormat(() => props.user.created_at, "DD/MM/YYYY, HH:mm")

const ctx = inject(UserDialogCtxKey)!
const api = useApi()
const toast = useToast()

// credit
const { execute: handleCredit, isLoading: isCreditLoading } = useAsyncState(
  async (body: HandlersCreditRequestInput) => {
    const { data } = await api.admin.creditUser(props.user.id!.toString(), body)
    toast.success({ message: data.message })
  },
  null,
  { immediate: false }
)

async function onCredit() {
  const result = await ctx.credit(props.user)
  if (result) {
    handleCredit(undefined, result)
  }
}

// drain
const { execute: handleDrain, isLoading: isDrainLoading } = useAsyncState(
  async () => {
    const { data } = await api.admin.drainUser(props.user.id!.toString())
    toast.success({ message: data.message })
  },
  null,
  { immediate: false }
)

async function onDrain() {
  const result = await ctx.drain(props.user)
  if (result) {
    await handleDrain()
  }
}

// remove
const { execute: handleRemove, isLoading: isRemoveLoading } = useAsyncState(
  async () => {
    const { data } = await api.admin.deleteUser(props.user.id!.toString())
    toast.success({ message: data.message })
    emit("remove")
  },
  null,
  { immediate: false }
)

async function onRemove() {
  const result = await ctx.remove(props.user)
  if (result) {
    await handleRemove()
  }
}
</script>
