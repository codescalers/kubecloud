<template>
  <AuthLayout
    title="Reset Your Access"
    quote="Security is built in — even when you forget a password."
    subtitle="Recover access to your account and continue building on decentralized cloud infrastructure."
    reversed
  >
    <v-tabs-window :model-value="activeTab" class="w-100 h-100">
      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668" class="w-100">
            <ForgotPasswordForm
              @back-to-login="activeTab = 0"
              @reset-password="
                /* prettier-ignore */
                activeTab = 1;
                email = $event
              "
            />
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668" class="w-100">
            <VerifyForm />
            <!-- <VerifyResetPassword :email="email" @verify="activeTab = 2" /> -->
          </v-card>
        </div>
      </v-tabs-window-item>

      <v-tabs-window-item class="h-100">
        <div class="form-container">
          <v-card max-width="668" class="w-100">
            <ResetPasswordForm :email="email" />
          </v-card>
        </div>
      </v-tabs-window-item>
    </v-tabs-window>
  </AuthLayout>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "non-auth" })

const activeTab = ref(2)
const email = ref("")
</script>
