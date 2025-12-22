<template>
  <div>
    <h3 class="text-h4 text-center mb-4">Create Account</h3>

    <p class="mb-10 text-center">Join Mycelium Cloud and start your journey</p>

    <v-form v-model="valid" @submit.prevent="register()">
      <v-row>
        <v-col cols="6">
          <v-text-field
            v-model.trim="username"
            variant="outlined"
            prepend-inner-icon="mdi-account"
            label="Username"
            placeholder="Enter your username"
            autofocus
            :rules="[
              (v) => !!v || 'Username is required',
              (v) => v.length >= 3 || 'Username must be at least 3 characters',
              (v) => v.length <= 12 || 'Username must be less than 12 characters',
            ]"
          />
        </v-col>

        <v-col cols="6">
          <v-text-field
            v-model.trim="email"
            prepend-inner-icon="mdi-email"
            variant="outlined"
            label="Email Address"
            placeholder="Enter your email address"
            :rules="[
              (v) => !!v || 'Email address is required',
              (v) => (v.includes('@') && v.includes('.')) || 'Email address must be valid',
            ]"
          />
        </v-col>

        <v-col cols="12">
          <PasswordInput v-model="password" />
        </v-col>
      </v-row>

      <v-btn
        type="submit"
        block
        size="large"
        text="Create Account"
        prepend-icon="mdi-account-plus"
        variant="outlined"
        class="mt-6"
        :disabled="!valid"
        :loading="isLoading"
      />
    </v-form>
    <!-- d-flex justify-center align-center  -->
    <div class="text-caption my-4">
      <span>By creating an account you agree to our</span>&nbsp;
      <a target="_blank" href="/terms-and-conditions" class="text-link">Terms & Conditions</a>&nbsp;
      <span>and</span>&nbsp;
      <a target="_blank" href="/privacy-policy" class="text-link">Privacy Policy</a>&nbsp;.
    </div>
  </div>
</template>

<script lang="ts" setup>
import type { HandlersRegisterInput } from "~/generated/api"

const props = defineProps<{ modelValue: HandlersRegisterInput | null }>()
const emit = defineEmits<{ (e: "update:model-value", value: HandlersRegisterInput | null): void }>()

const api = useApi()

const valid = ref(false)

const username = ref("")
const email = ref("")
const password = ref("")

const { execute: register, isLoading } = useAsyncState(
  async () => {
    if (props.modelValue !== null) {
      emit("update:model-value", null)
    }

    const registerBody = {
      name: username.value,
      email: email.value,
      password: password.value,
      confirm_password: password.value,
    }

    const { data: d1 } = await api.users.registerUser(registerBody, {
      unauthenticated: true,
    })
    const completed = await api.helpers.awaitWorkflowCompletion(d1.data?.workflow_id ?? "")

    if (completed) {
      emit("update:model-value", registerBody)
    }
  },
  null,
  { immediate: false }
)
</script>
