<template>
  <div>
    <h3 class="text-h4 text-center mb-4">Forgot Password</h3>

    <p class="mb-10 text-center">Enter your email to receive a reset code.</p>

    <v-form v-model="valid" @submit.prevent="forgotPassword()">
      <v-text-field
        v-model.trim="email"
        variant="outlined"
        prepend-inner-icon="mdi-email"
        label="Email Address"
        placeholder="Enter your email address"
        autofocus
        :rules="[
          (v) => !!v || 'Email address is required',
          (v) => (v.includes('@') && v.includes('.')) || 'Email address must be valid',
        ]"
      />

      <v-btn
        type="submit"
        block
        size="large"
        text="Reset Password"
        variant="outlined"
        class="mt-6"
        :disabled="!valid"
        :loading="isLoading"
      />
    </v-form>

    <div class="d-flex justify-center align-center mt-4">
      <v-btn
        prepend-icon="mdi-keyboard-backspace"
        flat
        text="Back To Login"
        @click="$emit('back-to-login')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
const emit = defineEmits<{
  (e: "back-to-login"): void
  (e: "reset-password", email: string): void
}>()

const api = useApi()
const valid = ref(false)
const email = ref("")

const { execute: forgotPassword, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.forgotPassword(
      { email: email.value },
      { unauthenticated: true, notify: true }
    )

    if (data.status === 200) {
      emit("reset-password", email.value)
    }
  },
  null,
  { immediate: false }
)
</script>
