<template>
  <div>
    <h3 class="text-h5 text-center mb-4">
      Please check your email
    </h3>
    <p class="text-body-2 text-center mb-4">
      We've sent a code to
      <span class="font-weight-bold">{{ email }}</span>
    </p>

    <v-form @submit.prevent>
      <v-otp-input
        v-model="otp"
        length="4"
        variant="outlined"
        autofocus
        :disabled="isLoading"
      />

      <v-btn
        type="submit"
        block
        size="large"
        text="Verify"
        variant="outlined"
        class="mt-4"
        prepend-icon="mdi-check-circle"
        :disabled="otp.length !== 4"
        :loading="isLoading"
        @click="verifyCode()"
      />
    </v-form>

    <div class="d-flex justify-center align-center mt-4">
      <p>
        Didn't receive the code?
        <v-btn text="Resend Code" variant="outlined" size="small" />
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ email: string }>()
const emit = defineEmits<{ (e: "verify"): void }>()

const api = useApi()
const otp = ref("")

const { accessToken, refreshToken } = useTokens()

const { execute: verifyCode, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.verifyForgotPasswordCode(
      { code: +otp.value, email: props.email },
      { unauthenticated: true },
    )

    accessToken.value = data.data?.access_token ?? ""
    refreshToken.value = data.data?.refresh_token ?? ""
    await nextTick() // make sure the tokens are updated

    emit("verify")
  },
  null,
  { immediate: false },
)
</script>
