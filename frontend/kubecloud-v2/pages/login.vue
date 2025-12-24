<template>
  <AuthLayout :cols="4" prepend-workart>
    <v-tabs-window :model-value="activeTab" class="w-100">
      <v-tabs-window-item>
        <LoginForm @forgot-password="activeTab = 1" />
      </v-tabs-window-item>

      <v-tabs-window-item>
        <ForgotPasswordForm
          @back-to-login="activeTab = 0"
          @reset-password="
            /* prettier-ignore */
            activeTab = 2;
            email = $event
          "
        />
      </v-tabs-window-item>

      <v-tabs-window-item>
        <VerifyResetPassword :email="email" @verify="activeTab = 3" />
      </v-tabs-window-item>

      <v-tabs-window-item>
        <ResetPasswordForm :email="email" />
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "non-auth" })

const activeTab = ref(3)
const email = ref("")
</script>
