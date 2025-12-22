<template>
  <v-row>
    <v-col :cols="block ? 12 : 6">
      <v-text-field
        v-model.trim="password"
        variant="outlined"
        label="Password"
        placeholder="Enter your password"
        prepend-inner-icon="mdi-lock"
        :rules="[
          (v) =>
            passwordRules.every((rule) => rule.rule(v)) || 'Please meet all password requirements',
        ]"
        :type="showPassword ? 'text' : 'password'"
        @input="confirmPassword && confirmPasswordRef?.validate()"
      >
        <template #append-inner>
          <v-icon
            :icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            tabindex="-1"
            @click="showPassword = !showPassword"
          />
        </template>
      </v-text-field>
    </v-col>

    <v-col cols="12" :order="block ? undefined : 'last'">
      <p class="text-body-2 opacity-70 mb-2">
        Password must contain at least 8 characters, including:
      </p>

      <v-list-item v-for="{ label, rule } in passwordRules" :key="label" density="compact">
        <template #prepend>
          <v-avatar :color="rule(password) ? 'success' : 'error'" variant="tonal" size="20">
            <v-icon v-if="rule(password)" icon="mdi-check" size="14" />
            <v-icon v-else icon="mdi-close" size="14" />
          </v-avatar>
        </template>

        <v-list-item-title class="text-body-2 opacity-60">
          {{ label }}
        </v-list-item-title>
      </v-list-item>
    </v-col>

    <v-col :cols="block ? 12 : 6">
      <v-text-field
        ref="confirmPasswordRef"
        v-model.trim="confirmPassword"
        variant="outlined"
        label="Confirm Password"
        prepend-inner-icon="mdi-lock-check"
        placeholder="Confirm your password"
        :rules="[
          (v) => !!v || 'Confirm password is required',
          (v) => v === password || 'Passwords do not match',
        ]"
        :type="showConfirmPassword ? 'text' : 'password'"
      >
        <template #append-inner>
          <v-icon
            :icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
            tabindex="-1"
            @click="showConfirmPassword = !showConfirmPassword"
          />
        </template>
      </v-text-field>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import type { VTextField } from "vuetify/components/VTextField"

defineProps<{ modelValue: string; block?: boolean }>()
const emit = defineEmits<{ (e: "update:model-value", value: string): void }>()

const confirmPasswordRef = ref<VTextField>()

const showPassword = ref(false)
const password = ref("")
watch(password, () => emit("update:model-value", password.value))

const showConfirmPassword = ref(false)
const confirmPassword = ref("")

// prettier-ignore
const passwordRules = markRaw([
  { label: "One uppercase letter (A-Z)", rule: (v: string) => v.match(/[A-Z]/) !== null },
  { label: "One lowercase letter (a-z)", rule: (v: string) => v.match(/[a-z]/) !== null },
  { label: "One number (0-9)", rule: (v: string) => v.match(/[0-9]/) !== null },
  { label: "One special character (!@#$%^&*)", rule: (v: string) => v.match(/[!@#$%^&*]/) !== null },
  { label: "At least 8 characters", rule: (v: string) => v.length >= 8 },
])
</script>
