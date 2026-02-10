<template>
  <AuthLayout
    title="Reset Your Access"
    quote="Security is built in — even when you forget a password."
    subtitle="Recover access to your account and continue building on decentralized cloud infrastructure."
    reversed
  >
    <v-tabs-window :model-value="activeTab" class="w-100 h-100">
      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668" class="w-100">
            <ForgotPasswordForm
              @reset="
                email = $event;
                activeTab = 1;
              "
            />
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668" class="w-100">
            <!-- <VerifyForm /> -->
            <VerifyForm
              :email="email"
              :active="activeTab === 1"
              back-label="Back To Forgot Password"
              :sending="isLoading"
              :verify-fn="verifyCode"
              @resend="sendVerificationCode()"
              @back="
                email = '';
                activeTab = 0;
              "
            />
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668" class="w-100">
            <ResetPasswordForm
              :email="email"
              :access-token="accessToken"
              :refresh-token="refreshToken"
            />
          </v-card>
        </div>
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "non-auth" })

const activeTab = ref(0)
const email = ref("")
const accessToken = ref("")
const refreshToken = ref("")

const api = useApi()
const { execute: sendVerificationCode, isLoading } = useAsyncState(
  () => api.users.forgotPassword({ email: email.value }, { unauthenticated: true }),
  null,
  { immediate: $meta.client },
)

async function verifyCode(otp: string) {
  const { data } = await api.users.verifyForgotPasswordCode(
    { code: +otp, email: email.value },
    { unauthenticated: true },
  )

  accessToken.value = data.data?.access_token ?? ""
  refreshToken.value = data.data?.refresh_token ?? ""
  await nextTick() // make sure the tokens are updated
  activeTab.value = 2
}
</script>
