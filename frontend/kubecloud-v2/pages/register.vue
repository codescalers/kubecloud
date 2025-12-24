<template>
  <AuthLayout :cols="6">
    <v-tabs-window :model-value="activeTab" class="w-100 h-100">
      <v-tabs-window-item class="h-100">
        <div class="pa-10 d-flex justify-center align-center h-100 w-100">
          <v-card
            max-width="668"
            class="mx-auto px-10 py-8 border w-100"
            :style="{ borderRadius: '18px' }"
          >
            <RegisterForm v-model="registerBody" />
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="pa-10 d-flex justify-center align-center h-100 w-100">
          <v-card
            max-width="668"
            class="mx-auto px-10 py-8 border w-100"
            :style="{ borderRadius: '18px' }"
          >
            <VerifyForm v-model="registerBody" />
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
</script>
