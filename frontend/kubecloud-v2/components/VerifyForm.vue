<template>
  <div class="text-center">
    <h3 class="text-h4 font-weight-bold mb-3">Verify Your email</h3>

    <p class="text-subtitle-2 mb-9">
      <span class="opacity-70">
        Please check your email for the verification code. We've sent a code to
      </span>
      &nbsp; <span class="font-weight-bold opacity-90" v-text="email" />.
      <span class="opacity-70">Please enter the 4-6 digits code.</span>
    </p>

    <v-form @submit.prevent="verifyCode()">
      <v-text-field
        v-model.trim="otp"
        variant="outlined"
        autofocus
        max-width="150"
        placeholder="123456"
        class="mx-auto code-input"
        maxlength="6"
        :rules="[(v) => v.length > 3]"
      />

      <p class="mb-2 text-caption">
        <span class="opacity-70 d-inline-block mb-1">Didn't receive the code?</span>&nbsp;
        <span v-if="isActive" class="font-weight-bold opacity-50">
          00:{{ remaining < 10 ? "0" : "" }}{{ remaining }}
        </span>
        <v-btn
          v-else
          type="button"
          class="font-weight-bold"
          text="Resend"
          variant="text"
          size="x-small"
          :disabled="isActive"
          @click="start(60)"
        />
      </p>

      <v-btn
        type="submit"
        block
        size="x-large"
        class="btn-form"
        text="Verify Email"
        prepend-icon="mdi-check-decagram-outline"
        variant="outlined"
        :disabled="otp.length < 4"
        :loading="isLoading"
      />
    </v-form>

    <div class="d-flex justify-center align-center">
      <v-btn
        class="mt-4 text-caption font-weight-bold"
        prepend-icon="mdi-arrow-left-thin"
        text="Back To Register"
        variant="plain"
        @click="$emit('update:model-value', null)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "~/generated/api"

const props = defineProps<{ modelValue: HandlersRegisterInput | null }>()
defineEmits<{ (e: "update:model-value", value: HandlersRegisterInput | null): void }>()

const router = useRouter()
const api = useApi()
const email = computed(() => {
  const e = props.modelValue?.email
  if (!e) {
    return ""
  }

  const parts = e.split("@")
  if (parts.length !== 2) {
    return e
  }

  return `${parts[0]!.slice(0, 2)}***@${parts[1]}`
})

const otp = ref("")

const { accessToken, refreshToken } = useTokens()

const { isActive, remaining, start } = useCountdown(60, { immediate: true })

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

<style lang="scss">
.code-input {
  input {
    text-align: center !important;
  }
}
</style>
