<template>
  <AuthLayout :cols="5">
    <v-tabs-window :model-value="activeTab" class="w-100">
      <v-tabs-window-item>
        <RegisterForm v-model="registerBody" />
      </v-tabs-window-item>

      <v-tabs-window-item>
        <VerifyForm v-model="registerBody" />
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "~/generated/api"

const registerBody = ref<HandlersRegisterInput | null>(null)
const activeTab = computed(() => (registerBody.value ? 1 : 0))
</script>
