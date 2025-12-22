<template>
  <AuthLayout :cols="4" prepend-workart>
    <v-tabs-window :model-value="activeTab" class="w-100">
      <v-tabs-window-item>
        <LoginForm @forgot-password="forgotPassword = true" />
      </v-tabs-window-item>

      <v-tabs-window-item>
        <ForgotPasswordForm @back-to-login="forgotPassword = false" />
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
const forgotPassword = ref(false)
const activeTab = computed(() => (forgotPassword.value ? 1 : 0))
</script>
