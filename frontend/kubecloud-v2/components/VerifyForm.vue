<template>
  <div>
    <h3 class="text-h5 text-center mb-4">Please check your email</h3>
    <p class="text-body-2 text-center mb-4">
      We've sent a code to
      <span class="font-weight-bold">{{ modelValue?.email }}</span>
    </p>

    <v-form @submit.prevent="verifyCode()">
      <v-otp-input v-model="otp" length="4" variant="outlined" autofocus :disabled="isLoading" />

      <v-btn
        type="submit"
        block
        size="x-large"
        class="btn-form"
        text="Verify"
        prepend-icon="mdi-check-circle-outline"
        variant="outlined"
        :disabled="otp.length !== 4"
        :loading="isLoading"
      />
    </v-form>

    <div class="d-flex justify-center align-center mt-4">
      <p>
        Didn't receive the code?
        <v-btn text="Resend Code" variant="outlined" size="small" />
      </p>
    </div>

    <v-btn text="Another Account?" variant="text" @click="$emit('update:model-value', null)" />
  </div>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "~/generated/api"

const props = defineProps<{ modelValue: HandlersRegisterInput | null }>()
defineEmits<{ (e: "update:model-value", value: HandlersRegisterInput | null): void }>()

const router = useRouter()
const api = useApi()
const otp = ref("")

const { accessToken, refreshToken } = useTokens()

const { execute: verifyCode, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.verifyRegisterCode(
      {
        code: +otp.value,
        email: props.modelValue?.email ?? "",
      },
      { unauthenticated: true }
    )

    const completed = await api.helpers.awaitWorkflowCompletion(data.data?.workflow_id ?? "")
    if (!completed) {
      return console.log("failed")
    }

    const { data: loginData } = await api.users.loginUser(
      {
        email: props.modelValue?.email ?? "",
        password: props.modelValue?.password ?? "",
      },
      { unauthenticated: true }
    )

    accessToken.value = loginData.data?.access_token ?? ""
    refreshToken.value = loginData.data?.refresh_token ?? ""
    await nextTick() // make sure the tokens are updated

    router.push("/dashboard")
  },
  null,
  { immediate: false }
)
</script>
