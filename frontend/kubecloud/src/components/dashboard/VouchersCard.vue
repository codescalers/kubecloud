<template>
  <v-card color="surface-variant" class="pa-6" style="max-width: 50rem;">
    <div class="mb-6">
      <h3 class="text-h5 font-weight-bold mb-2">Redeem Voucher</h3>
      <p class="text-body-1 text-primary">Add credits to your balance using a voucher code</p>
    </div>

    <div class="d-flex flex-column ga-4">
      <v-text-field
        v-model="code"
        label="Voucher Code"
        :disabled="loading"
        @keyup.enter="onRedeem"
        variant="outlined"
        color="primary"
        hide-details="auto"
        density="comfortable"
        :append-inner-icon="code ? 'mdi-close' : ''"
        @click:append-inner="code = ''"
      />
      <v-btn
        color="primary"
        variant="elevated"
        size="large"
        :loading="loading"
        :disabled="loading || !code.trim()"
        @click="onRedeem"
        prepend-icon="mdi-gift"
      >
        Redeem
      </v-btn>
      <div v-if="successMessage" class="text-success font-weight-medium">{{ successMessage }}</div>
      <div v-if="errorMessage" class="text-error font-weight-medium">{{ errorMessage }}</div>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { userService } from '../../utils/userService'
import { useUserStore } from '../../stores/user'


const code = ref('')
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')
const userStore = useUserStore()

async function onRedeem() {
  if (!code.value.trim()) return
  loading.value = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await userService.redeemVoucher(code.value.trim())
    code.value = ''
  } catch (err: any) {
    console.error(err)
  } finally {
    loading.value = false
  }
  await userStore.updateNetBalance()
}

</script>


