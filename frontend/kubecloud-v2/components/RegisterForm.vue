<template>
  <div>
    <h1 class="text-h4 font-weight-bold mb-9">Create Account</h1>

    <v-form v-model="valid" :disabled="isLoading" @submit.prevent="register()">
      <v-row>
        <v-col cols="12" sm="6" md="12" lg="6">
          <v-text-field
            v-model.trim="username"
            variant="outlined"
            prepend-inner-icon="mdi-account-outline"
            label="Username"
            placeholder="Enter your username"
            :rules="[
              (v) => !!v || 'Username is required',
              (v) => v.length >= 3 || 'Username must be at least 3 characters',
              (v) => v.length <= 12 || 'Username must be less than 12 characters',
            ]"
            autofocus
          />
        </v-col>

        <v-col cols="12" sm="6" md="12" lg="6">
          <v-text-field
            ref="emailRef"
            v-model.trim="email"
            prepend-inner-icon="mdi-email-outline"
            variant="outlined"
            label="Email Address"
            placeholder="Enter your email address"
            :rules="[
              (v) => !!v || 'Email address is required',
              (v) => isEmail(v) || 'Email address must be valid',
            ]"
            :error="emailInUse || undefined"
            :error-messages="emailInUse ? ['Email already in use'] : undefined"
            @input="emailInUse && (emailInUse = false)"
          />
        </v-col>

        <v-col cols="12">
          <PasswordInput v-model="password" />
        </v-col>
      </v-row>

      <div class="d-flex justify-center pr-2">
        <v-checkbox v-model="termsAccepted" class="mt-5 mb-0" hide-details color="primary">
          <template #label>
            <div class="text-caption">
              <span class="opacity-50">By creating an account you agree to our</span>&nbsp;
              <a target="_blank" href="/terms-and-conditions" class="text-link"
                >Terms & Conditions</a
              >&nbsp; <span class="opacity-50">and</span>&nbsp;
              <a target="_blank" href="/privacy-policy" class="text-link">Privacy Policy</a>
            </div>
          </template>
        </v-checkbox>
      </div>

      <v-btn
        type="submit"
        block
        size="x-large"
        class="btn-form"
        text="Create Account"
        prepend-icon="mdi-account-plus-outline"
        variant="outlined"
        :disabled="!valid || !termsAccepted"
        :loading="isLoading"
      />
    </v-form>
  </div>
</template>

<script lang="ts" setup>
import { AxiosError } from "axios"
import isEmail from "validator/es/lib/isEmail"
import type { HandlersRegisterInput } from "~/generated/api"

const props = defineProps<{ modelValue: HandlersRegisterInput | null }>()
const emit = defineEmits<{ (e: "update:model-value", value: HandlersRegisterInput | null): void }>()

const api = useApi()

const valid = ref(false)
const emailInUse = ref(false)

const username = ref("")
const email = ref("")
const password = ref("")
const termsAccepted = ref(false)

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

    const { data: d1 } = await api.users.registerUser(registerBody, { unauthenticated: true })
    const completed = await api.helpers.awaitWorkflowCompletion(d1.data?.workflow_id ?? "")

    if (completed) {
      emit("update:model-value", registerBody)
    }
  },
  null,
  {
    immediate: false,
    onError(err: unknown) {
      if (err instanceof AxiosError) {
        if (err.response?.status === 409) {
          emailInUse.value = true
        }
      }
    },
  }
)
</script>
