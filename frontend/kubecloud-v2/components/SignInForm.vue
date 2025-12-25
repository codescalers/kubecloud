<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-center">Welcome Back!</h1>

    <p class="text-subtitle-2 opacity-70 text-center mt-4 mb-8">
      Sign in to your Mycelium Cloud account
    </p>

    <v-form v-model="valid" @submit.prevent="login()">
      <v-text-field
        v-model="email"
        prepend-inner-icon="mdi-email-outline"
        variant="outlined"
        label="Email Address"
        :rules="[(v) => !!v, (v) => isEmail(v)]"
      />

      <v-text-field
        v-model="password"
        variant="outlined"
        prepend-inner-icon="mdi-lock-outline"
        label="Password"
        :rules="[(v) => !!v]"
        :type="showPassword ? 'text' : 'password'"
        hide-details
      >
        <template #append-inner>
          <v-icon
            :icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
            tabindex="-1"
            @click="showPassword = !showPassword"
          />
        </template>
      </v-text-field>

      <NuxtLink to="/forgot-password" class="text-link text-caption d-inline-block mt-2 mb-8">
        Forgot Password?
      </NuxtLink>

      <v-btn
        type="submit"
        block
        size="x-large"
        class="btn-form"
        text="Sign In"
        prepend-icon="mdi-login"
        variant="outlined"
        :disabled="!valid"
        :loading="isLoading"
      />
    </v-form>
  </div>
</template>

<script setup lang="ts">
import { AxiosError } from "axios"
import { isEmail } from "validator"

defineEmits<{ (e: "forgot-password"): void }>()

const valid = ref(false)
const showPassword = ref(false)
const email = ref("")
const password = ref("")

const { accessToken, refreshToken } = useTokens()
const router = useRouter()
const toast = useToast()

const api = useApi()
const { isLoading, execute: login } = useAsyncState(
  async () => {
    const { data } = await api.users.loginUser(
      { email: email.value, password: password.value },
      { unauthenticated: true }
    )

    accessToken.value = data.data?.access_token ?? ""
    refreshToken.value = data.data?.refresh_token ?? ""

    await nextTick()
    toast.success({ message: "You’ve logged in successfully" })
    router.push("/dashboard")
  },
  null,
  {
    immediate: false,
    onError(e) {
      if (e instanceof AxiosError) {
        toast.error({ message: e.response?.data?.message ?? "An unknown error occurred" })
      }
    },
  }
)
</script>
