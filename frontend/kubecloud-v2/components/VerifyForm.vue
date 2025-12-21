<template>
  <v-container max-width="400" class="h-100 d-flex justify-center align-center">
    <v-card>
      <v-card-text>
        <h3 class="text-h5 text-center mb-4">Please check your email</h3>
        <p class="text-body-2 text-center mb-4">
          We've sent a code to
          <span class="font-weight-bold">{{ modelValue?.email }}</span>
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

        <v-btn text="Another Account?" variant="text" @click="$emit('update:model-value', null)" />
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "../generated/api"

const props = defineProps<{ modelValue: HandlersRegisterInput | null }>()
defineEmits<{ (e: "update:model-value", value: HandlersRegisterInput | null): void }>()

const router = useRouter()
const api = useApi()
const otp = ref("")

const accessToken = useLocalStorage<string>("accessToken", "", { writeDefaults: false })
const refreshToken = useLocalStorage<string>("refreshToken", "", { writeDefaults: false })

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

    router.push("/dashboard")
  },
  null,
  { immediate: false }
)
</script>
