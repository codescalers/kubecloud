<template>
  <div class="text-center">
    <h3 class="text-h4 font-weight-bold mb-3">
      Verify Your email
    </h3>

    <p class="text-subtitle-2 mb-9">
      <span class="opacity-70">
        Please check your email for the verification code. We've sent a code to
      </span>
      <span
        class="font-weight-bold opacity-90"
        v-text="email ? `${email.slice(0, 2)}***@${email.split('@')[1]}` : 'Your Email Address'"
      />.
      <span class="opacity-70">Please enter the 4-6 digits code.</span>
    </p>

    <v-form :disabled="verifying || sending" @submit.prevent="verify()">
      <v-text-field
        v-model.trim="otp"
        variant="outlined"
        autofocus
        max-width="150"
        placeholder="123456"
        class="mx-auto code-input"
        maxlength="6"
        :rules="[(v) => v.length > 3]"
        :error="error || undefined"
        @input="error && (error = false)"
      />

      <p class="mb-2 text-caption">
        <span class="opacity-70 d-inline-block mb-1">Didn't receive the code?</span>&nbsp;
        <span
          v-if="isActive && !sending"
          class="font-weight-bold opacity-50 text-left d-inline-block pl-4"
          :style="{ width: '67px' }"
        >
          00:{{ remaining < 10 ? "0" : "" }}{{ remaining }}
        </span>
        <v-btn
          v-else
          type="button"
          class="font-weight-bold"
          text="Resend"
          variant="text"
          size="x-small"
          :loading="sending"
          :disabled="verifying"
          @click="
            // prettier-ignore
            $emit('resend');
            start(30)
          "
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
        :loading="verifying"
      />
    </v-form>

    <div class="d-flex justify-center align-center">
      <v-btn
        class="mt-4 text-caption font-weight-bold"
        prepend-icon="mdi-arrow-left-thin"
        :text="backLabel"
        variant="plain"
        :disabled="verifying || sending"
        @click="$emit('back')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { AxiosError } from "axios"

const props = defineProps<{
  email: string | undefined
  active: boolean
  backLabel: string
  sending: boolean
  verifyFn: (otp: string) => Promise<void>
}>()

defineEmits<{ (e: "resend" | "back"): void }>()

const otp = ref("")
const error = ref(false)

const { isActive, remaining, start } = useCountdown(30)
watch(
  () => props.active,
  (v) => {
    if (v) {
      start(30)
      otp.value = ""
    }
  },
  { immediate: true },
)

const toast = useToast()
const { execute: verify, isLoading: verifying } = useAsyncState(
  () => props.verifyFn(otp.value),
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

<style lang="scss">
.code-input {
  input {
    text-align: center !important;
  }
}
</style>
