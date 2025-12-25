<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-center">Reset Password</h1>

    <p class="text-subtitle-2 opacity-70 text-center mt-2 mb-10">Enter your new password below</p>

    <v-form v-model="valid" @submit.prevent="resetPassword()">
      <PasswordInput v-model="password" block />

      <v-btn
        type="submit"
        block
        size="x-large"
        class="btn-form mt-5"
        text="Update Password"
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

const props = defineProps<{ email: string; accessToken: string; refreshToken: string }>()

const password = ref("")
const valid = ref(false)
const router = useRouter()
const api = useApi()

const tokens = useTokens()

const toast = useToast()
const { execute: resetPassword, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.changePassword(
      {
        email: props.email,
        password: password.value,
        confirm_password: password.value,
      },
      { unauthenticated: true, headers: { Authorization: `Bearer ${props.accessToken}` } }
    )

    if (data.status === 200) {
      tokens.accessToken.value = props.accessToken
      tokens.refreshToken.value = props.refreshToken

      await nextTick()

      toast.success({ message: "Password updated successfully" })
      router.push("/dashboard")
    }
  },
  null,
  {
    immediate: false,
    onError(e: unknown) {
      if (e instanceof AxiosError) {
        toast.error({ message: e.response?.data?.message ?? "An unknown error occurred" })
      }
    },
  }
)
</script>
