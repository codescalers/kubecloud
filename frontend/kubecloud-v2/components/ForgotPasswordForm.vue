<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-center">
      Forgot Password
    </h1>

    <p class="text-subtitle-2 opacity-70 text-center mt-2 mb-8">
      Enter your email to receive a reset code
    </p>

    <v-form v-model="valid" @submit.prevent="forgotPassword()">
      <v-text-field
        v-model.trim="email"
        variant="outlined"
        prepend-inner-icon="mdi-email"
        label="Email Address"
        placeholder="Enter your email address"
        autofocus
        hide-details
        :rules="[
          (v) => !!v || 'Email address is required',
          (v) => (v.includes('@') && v.includes('.')) || 'Email address must be valid',
        ]"
        :error="error || undefined"
        @input="error && (error = false)"
      />

      <v-btn
        type="submit"
        block
        size="x-large"
        class="btn-form mt-5"
        text="Reset Password"
        prepend-icon="mdi-lock-reset"
        variant="outlined"
        :disabled="!valid"
        :loading="isLoading"
      />
    </v-form>
  </div>
</template>

<script setup lang="ts">
import { AxiosError } from "axios"

const emit = defineEmits<{ (e: "reset", email: string): void }>()

const api = useApi()
const valid = ref(false)
const email = ref("")
const error = ref(false)

const toast = useToast()
const { execute: forgotPassword, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.forgotPassword(
      { email: email.value },
      { unauthenticated: true },
    )

    if (data.status === 200) {
      emit("reset", email.value)
    }
  },
  null,
  {
    immediate: false,
    onError(e: unknown) {
      if (e instanceof AxiosError) {
        error.value = true
        toast.error({ message: e.response?.data?.message ?? "An unknown error occurred" })
      }
    },
  },
)
</script>
