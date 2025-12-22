<template>
  <div>
    <h3 class="text-h5 text-center mb-4">Reset Password</h3>
    <p class="text-body-2 text-center mb-4">Enter your new password below.</p>

    <v-form v-model="valid" @submit.prevent="resetPassword()">
      <PasswordInput v-model="password" block />

      <v-btn
        type="submit"
        block
        size="large"
        text="Verify"
        variant="outlined"
        class="mt-4"
        prepend-icon="mdi-check-circle"
        :disabled="!valid"
        :loading="isLoading"
      />
    </v-form>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ email: string }>()

const password = ref("")
const valid = ref(false)
const router = useRouter()
const api = useApi()
const { execute: resetPassword, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.changePassword(
      {
        email: props.email,
        password: password.value,
        confirm_password: password.value,
      },
      { notify: true }
    )

    if (data.status === 200) {
      router.push("/dashboard")
    }
  },
  null,
  { immediate: false }
)
</script>
