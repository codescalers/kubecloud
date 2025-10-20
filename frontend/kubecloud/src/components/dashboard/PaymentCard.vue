<template>
  <div class="dashboard-card payment-card spacious">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Add Funds</h3>
      <p class="dashboard-card-subtitle">Add funds to your account balance</p>
    </div>
    <div class="dashboard-card-content">
      <div class="balance-container">
      <div class="balance-row">
        <span>Current Balance:</span>
        <span class="balance-value">${{ userStore.netBalance.toFixed(2) }}</span>
      </div>
      </div>
      <div class="amount-row">
        <span>Amount:</span>
        <div class="amount-options">
          <button
            v-for="preset in presets"
            :key="preset"
            :class="{ selected: typeof amount === 'number' && amount === preset }"
            @click="selectAmount(preset)"
            type="button"
          >{{ preset }}</button>
          <template v-if="typeof amount !== 'string' || amount !== 'custom'">
            <button
              :class="{ selected: typeof amount === 'string' && amount === 'custom' }"
              @click="selectAmount('custom')"
              type="button"
            >Custom</button>
          </template>
          <input
            v-if="typeof amount === 'string' && amount === 'custom'"
            type="number"
            min="1"
            class="amount-input"
            v-model.number="customAmount"
            placeholder="Custom"
            @focus="selectAmount('custom')"
          />
        </div>
      </div>
      <div class="card-details-row">
        <div id="stripe-card-element" class="stripe-card-element"></div>
      </div>
      <v-btn
        class="action-btn"
        color="primary"
        :loading="loading"
        :disabled="loading || !isFormValid"
        @click="chargeBalance"
        prepend-icon="mdi-credit-card-plus"
      >
        Charge Balance
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, type Ref, onMounted } from 'vue'
import { useUserStore } from '../../stores/user'
import { userService } from '../../utils/userService'
import type { StripeElements, StripeCardElement, StripeElementsOptions } from '@stripe/stripe-js'
import { stripeService } from '../../utils/stripeService'

const userStore = useUserStore()

const presets = [5, 10, 20, 50]
const amount: Ref<number | 'custom'> = ref(5)
const customAmount = ref<number | null>(null)
const loading = ref(false)
const cardComplete = ref(false)

// Stripe Elements
const stripe = ref<any>(null)
const elements = ref<StripeElements | null>(null)
const cardElement = ref<StripeCardElement | null>(null)
const stripeLoaded = ref(false)

onMounted(async () => {
  await stripeService.initialize()
  stripe.value = await stripeService.getStripe()
  elements.value = stripe.value.elements()
  const container = document.getElementById('stripe-card-element')
  if (elements.value && container) {
    cardElement.value = elements.value.create('card', {
      style: { base: { color: '#CBD5E1', fontFamily: 'Inter, sans-serif', fontSize: '16px' } },
      hidePostalCode: true
    })
    cardElement.value.mount('#stripe-card-element')
    cardElement.value.on('change', (event: any) => {
      cardComplete.value = !!event.complete
    })
    stripeLoaded.value = true
  }
})

const isFormValid = computed(() => {
  const selectedAmount = getSelectedAmount()
  return selectedAmount && selectedAmount > 0 && stripeLoaded.value && cardComplete.value
})

function selectAmount(val: number | 'custom') {
  amount.value = val
  if (val !== 'custom') customAmount.value = null
}

async function chargeBalance() {
  loading.value = true
  const selectedAmount = getSelectedAmount()
  if (!selectedAmount || !isFormValid.value) {
    loading.value = false
    return
  }
  try {
    // Create token with Stripe (for backend expecting 'tok_' id)
    const tokenId = await stripeService.createToken(cardElement.value)
    await userService.chargeBalance({
      card_type: 'card',
      payment_method_id: tokenId, // This is now a 'tok_' id
      amount: Number(selectedAmount)
    })
    await userStore.updateNetBalance()
    // Clear the form
    if (cardElement.value) cardElement.value.clear()
    amount.value = 5
    customAmount.value = null
  } catch (err: any) {
    console.error('Failed to charge balance:', err)
  } finally {
    loading.value = false
  }
}

function getSelectedAmount() {
  if (typeof amount.value === 'string' && amount.value === 'custom') {
    return customAmount.value
  }
  if (typeof amount.value === 'number') {
    return amount.value
  }
  return null
}
</script>

<style scoped>
.dashboard-card.payment-card.spacious {
  background: #181f35;
  border-radius: 1.25rem;
  border: 1.5px solid #334155;
  max-width: 50rem;
  margin: 0;
  padding: 2.2rem 2.5rem 2.2rem 2.5rem;
  box-shadow: 0 4px 32px 0 rgba(0,0,0,0.12);
  display: flex;
  flex-direction: column;
}
.balance-container {
  margin-bottom: 1.3rem;
  font-size: 1.1rem;
}
.balance-row {
  display: flex;
  gap: 2rem;
  justify-content: space-between;
  align-items: center;
}
.dashboard-card-header {
  margin-bottom: 1.5rem;
  width: 100%;
}
.dashboard-card-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.2rem;
}
.dashboard-card-subtitle {
  font-size: 1.05rem;
  color: #60a5fa;
  margin-bottom: 0.5rem;
}
.dashboard-card-content {
  width: 100%;
}

.balance-value {
  color: #10B981;
  font-weight: 700;
  font-size: 1.2rem;
}
.pending-balance-value {
  color: rgba(203, 213, 225, 0.5);
  font-weight: 700;
  font-size: 1rem;
}
.amount-row {
  display: flex;
  align-items: center;
  margin-bottom: 1.3rem;
  gap: 0.7rem;
}
.amount-options {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.amount-btn,
.amount-options button {
  background: #232f47;
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  padding: 0.5rem 1.3rem;
  font-size: 1.1rem;
  color: #CBD5E1;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}
.amount-options button.selected,
.amount-options button:focus {
  background: #60a5fa;
  color: #fff;
  border-color: #60a5fa;
}
.amount-input {
  width: 100px;
  padding: 0.5rem 0.7rem;
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  font-size: 1.1rem;
  margin-left: 0.3rem;
}
.card-details-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.3rem;
}
.stripe-card-element {
  background: #232f47;
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  padding: 0.7rem 1.1rem;
  font-size: 1.1rem;
  color: #CBD5E1;
  margin-bottom: 1.3rem;
  min-height: 44px;
  min-width: 29rem;
  display: block;
}
.pending-balance-text {
  font-size: 1rem;
  color: rgba(203, 213, 225, 0.8);
}
.action-btn {
  margin-top: 1.2rem;
  font-size: 1.1rem;
  border-radius: 0.7rem;
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
