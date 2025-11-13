<template>
  <div class="dashboard-card vouchers-card spacious">
    <div class="header-content">
      <h3 class="text-h5 font-weight-bold mb-1">Redeem Voucher</h3>
      <p class="text-body-1 dashboard-card-subtitle">Add credits to your balance using a voucher code</p>
    </div>
    <div class="dashboard-card-content">
      <v-text-field
        v-model="code"
        class="voucher-input-field"
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
        :loading="loading"
        :disabled="loading || !code.trim()"
        class="redeem-btn"
        @click="onRedeem"
        prepend-icon="mdi-gift"
      >
        Redeem
      </v-btn>
      <div v-if="successMessage" class="success-message mt-3">{{ successMessage }}</div>
      <div v-if="errorMessage" class="error-message mt-3">{{ errorMessage }}</div>
    </div>
  </div>
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

<style scoped>
.dashboard-card.vouchers-card.spacious {
  background: #181f35;
  border-radius: 1.25rem;
  border: 1.5px solid #334155;
  max-width: 50rem;
  margin-left: 0;
  padding: 2.2rem 2.5rem 2.2rem 2.5rem;
  box-shadow: 0 4px 32px 0 rgba(0,0,0,0.12);
  display: flex;
  flex-direction: column;
}
.header-content {
  margin-bottom: var(--space-6);
}
.dashboard-card-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.2rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.single-line {
  white-space: nowrap;
}
.dashboard-card-subtitle {
  font-size: 1.05rem;
  color: #60a5fa;
  margin-bottom: 0.5rem;
  flex-shrink: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dashboard-card-content {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
  align-items: stretch;
}
.voucher-input-field {
  width: 100%;
}
.redeem-btn {
  max-width: 12rem;
  font-size: 1.1rem;
  border-radius: 0.7rem;
  margin-top: 0.2rem;
}
.success-message {
  color: #10B981;
  font-weight: 500;
  font-size: 1rem;
}
.error-message {
  color: #EF4444;
  font-weight: 500;
  font-size: 1rem;
}
</style>
