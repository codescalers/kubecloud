<template>
  <v-row>
    <v-col cols="12" sm="6" md="12" lg="6">
      <v-text-field
        v-model.trim="password"
        variant="outlined"
        label="Password"
        placeholder="Enter your password"
        prepend-inner-icon="mdi-lock-outline"
        :rules="[
          (v) => passwordRules.slice(0, passwordRules.length - 1).every((rule) => rule.rule(v)),
        ]"
        :type="showPassword ? 'text' : 'password'"
        hide-details
        @input="confirmPassword && confirmPasswordRef?.validate()"
      >
        <template #append-inner>
          <v-icon
            :icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
            tabindex="-1"
            @click="showPassword = !showPassword"
          />
        </template>
      </v-text-field>
    </v-col>

    <v-col cols="12" order-sm="last" order-md="0" order-lg="last">
      <v-card flat class="bg-2 border pa-4" :style="{ borderRadius: '8px' }">
        <p class="mb-2" :style="{ fontSize: '12px' }">
          Password must contain at least 8 characters, including:
        </p>

        <v-list-item
          v-for="{ label, rule } in passwordRules"
          :key="label"
          class="pl-0"
          density="compact"
        >
          <template #prepend>
            <v-avatar color="transparent" size="20" :style="{ marginRight: '-6px' }">
              <v-icon v-if="rule(password)" size="small" icon="mdi-check" color="success" />
              <v-icon v-else icon="mdi-close" size="small" color="error" />
            </v-avatar>
          </template>

          <v-list-item-title class="opacity-50 text-caption">
            {{ label }}
          </v-list-item-title>
        </v-list-item>
      </v-card>
    </v-col>

    <v-col cols="12" sm="6" md="12" lg="6">
      <v-text-field
        ref="confirmPasswordRef"
        v-model.trim="confirmPassword"
        variant="outlined"
        label="Confirm Password"
        prepend-inner-icon="mdi-lock-check-outline"
        placeholder="Confirm your password"
        :rules="[(v) => !!v, (v) => v === password]"
        :type="showConfirmPassword ? 'text' : 'password'"
        hide-details
      >
        <template #append-inner>
          <v-icon
            :icon="showConfirmPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
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
  { label: "Passwords should match", rule: (v: string) => !!v && v === confirmPassword.value },
])
</script>
