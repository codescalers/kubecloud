<template>
  <AuthLayout
    title="Join Mycelium Cloud"
    quote="Start your journey into decentralized cloud infrastructure."
    subtitle="Create your account to access secure compute, scalable resources, and a next-generation cloud built without central points of failure."
  >
    <v-tabs-window :model-value="activeTab" class="w-100 h-100">
      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668">
            <RegisterForm v-model="registerBody" />
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668">
            <VerifyForm
              :email="registerBody?.email"
              :active="activeTab === 1"
              back-label="Back To Register"
              :sending="isLoading"
              :verify-fn="verifyCode"
              @resend="sendVerificationCode()"
              @back="registerBody = null"
            />
          </v-card>
        </div>
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "~/generated/api"

definePageMeta({ middleware: "non-auth" })

const registerBody = ref<HandlersRegisterInput | null>(null)
const activeTab = computed(() => (registerBody.value ? 1 : 0))

const api = useApi()
const { execute: sendVerificationCode, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.users.registerUser(registerBody.value!, { unauthenticated: true })
    await api.helpers.awaitWorkflowCompletion(data.data?.workflow_id ?? "")
  },
  null,
  { immediate: false }
)

const router = useRouter()
const { accessToken, refreshToken } = useTokens()

const toast = useToast()
async function verifyCode(otp: string) {
  const body = registerBody.value!

  const { data } = await api.users.verifyRegisterCode(
    { code: +otp, email: body.email },
    { unauthenticated: true }
  )

  const completed = await api.helpers.awaitWorkflowCompletion(data.data?.workflow_id ?? "")
  if (!completed) {
    return toast.error({ message: "Failed to verify email" })
  }

  const { data: loginData } = await api.users.loginUser(
    {
      email: body.email,
      password: body.password,
    },
    { unauthenticated: true }
  )

  accessToken.value = loginData.data?.access_token ?? ""
  refreshToken.value = loginData.data?.refresh_token ?? ""
  await nextTick() // make sure the tokens are updated

  toast.success({ message: "You’ve registered successfully" })
  router.push("/dashboard")
}
</script>
