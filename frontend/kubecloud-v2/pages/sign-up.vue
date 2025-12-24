<template>
  <AuthLayout
    title="Join Mycelium Cloud"
    quote="Start your journey into decentralized cloud infrastructure."
    subtitle="Create your account to access secure compute, scalable resources, and a next-generation cloud built without central points of failure."
  >
    <v-tabs-window :model-value="activeTab" class="w-100 h-100">
      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668">
            <RegisterForm v-model="registerBody" />
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668">
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
