<template>
  <v-row no-gutters class="h-100">
    <v-col cols="6">
      <v-tabs-window :model-value="activeTab" class="h-100">
        <v-tabs-window-item class="h-100">
          <RegisterForm v-model="registerBody" />
        </v-tabs-window-item>

        <v-tabs-window-item class="h-100">
          <VerifyForm v-model="registerBody" />
        </v-tabs-window-item>
      </v-tabs-window>
    </v-col>

    <v-col cols="6">
      <div class="h-100 w-100 bg-primary d-flex justify-center align-center">KubeCloud</div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import type { HandlersRegisterInput } from "../../generated/api"

const registerBody = ref<HandlersRegisterInput | null>(null)
const activeTab = computed(() => (registerBody.value ? 1 : 0))
</script>
