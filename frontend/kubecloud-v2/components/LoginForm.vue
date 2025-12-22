<template>
  <div>
    <h3 class="text-h4 text-center mb-4">Welcome Back!</h3>

    <p class="mb-10 text-center">Sign in to your Mycelium Cloud account</p>

    <v-form v-model="valid" @submit.prevent="login()">
      <v-text-field v-model="email" variant="outlined" label="Email Address" />
      <v-text-field v-model="password" variant="outlined" label="Password" type="password" />

      <!-- <v-btn
        type="button"
        text="Forgot Password?"
        variant="text"
        size="small"
        @click="$emit('forgot-password')"
      /> -->
      <a href="#!" class="text-link text-caption" @click.prevent="$emit('forgot-password')">
        Forgot Password?
      </a>

      <v-btn
        type="submit"
        block
        size="large"
        text="Login"
        variant="outlined"
        class="mt-4"
        :disabled="!valid"
        :loading="isLoading"
      />
    </v-form>
  </div>
</template>

<script setup lang="ts">
defineEmits<{ (e: "forgot-password"): void }>()

const valid = ref(false)
const email = ref("")
const password = ref("")

const { accessToken, refreshToken } = useTokens()
const router = useRouter()

const api = useApi()
const { isLoading, execute: login } = useAsyncState(
  async () => {
    const { data } = await api.users.loginUser({
      email: email.value,
      password: password.value,
    })

    accessToken.value = data.data?.access_token ?? ""
    refreshToken.value = data.data?.refresh_token ?? ""

    await nextTick()
    router.push("/dashboard")
  },
  null,
  { immediate: false }
)
</script>
