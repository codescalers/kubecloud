<template>
  <v-container max-width="400" class="h-100 d-flex justify-center align-center">
    <v-card>
      <v-card-text>
        <h3 class="text-h5 text-center mb-4">Please check your email</h3>
        <p class="text-body-2 text-center mb-4">
          We've sent a code to
          <span class="font-weight-bold">{{ registerationFormData?.email }}</span>
        </p>

        <v-form @submit.prevent>
          <v-otp-input v-model="otp" length="4" variant="outlined" autofocus disabled />

          <v-btn
            type="submit"
            block
            size="large"
            text="Verify"
            variant="outlined"
            class="mt-4"
            prepend-icon="mdi-check-circle"
          />
        </v-form>

        <div class="d-flex justify-center align-center mt-4">
          <p>
            Didn't receive the code?
            <v-btn text="Resend Code" variant="outlined" size="small" />
          </p>
        </div>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
const registerationFormData = useRegisterationFormData()

const otp = ref("")
</script>
