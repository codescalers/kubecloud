<template>
  <AuthLayout :cols="6">
    <v-tabs-window :model-value="activeTab" class="w-100">
      <v-tabs-window-item>
        <v-card
          max-width="668"
          class="mx-auto px-10 py-8 border"
          :style="{ borderRadius: '18px', backgroundColor: '#1821348C' }"
        >
          <RegisterForm v-model="registerBody" />
        </v-card>
      </v-tabs-window-item>

      <v-tabs-window-item>
        <VerifyForm v-model="registerBody" />
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "~/generated/api"

definePageMeta({ middleware: "non-auth" })

const registerBody = ref<HandlersRegisterInput | null>(null)
const activeTab = computed(() => (registerBody.value ? 1 : 0))
</script>
