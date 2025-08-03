<template>
  <v-card color="surface-variant" class="pa-6" style="max-width: 50rem;">
    <div class="mb-6">
      <h3 class="text-h5 font-weight-bold mb-2">Add Funds</h3>
      <p class="text-body-1 text-primary">Add funds to your account balance</p>
    </div>

    <div class="d-flex justify-space-between align-center mb-6">
      <span class="text-body-1">Current Balance:</span>
      <span class="text-h6 font-weight-bold text-success">${{ userStore.netBalance.toFixed(2) }}</span>
    </div>

    <div class="mb-6">
      <div class="text-body-1 mb-3">Amount:</div>
      <div class="d-flex flex-wrap ga-2 mb-4">
        <v-btn
          v-for="preset in presets"
          :key="preset"
          :variant="typeof amount === 'number' && amount === preset ? 'elevated' : 'outlined'"
          :color="typeof amount === 'number' && amount === preset ? 'primary' : 'default'"
          @click="selectAmount(preset)"
          size="small"
        >${{ preset }}</v-btn>
        <v-btn
          :variant="typeof amount === 'string' && amount === 'custom' ? 'elevated' : 'outlined'"
          :color="typeof amount === 'string' && amount === 'custom' ? 'primary' : 'default'"
          @click="selectAmount('custom')"
          size="small"
        >Custom</v-btn>
      </div>
      <v-text-field
        v-if="typeof amount === 'string' && amount === 'custom'"
        v-model.number="customAmount"
        type="number"
        label="Enter amount"
        prefix="$"
        :min="1"
        variant="outlined"
      />
    </div>

    <div class="mb-6">
      <div class="text-body-1 mb-3">Card Details:</div>
      <div id="stripe-card-element" class="stripe-card-element pa-3" style="border: 1px solid rgba(96, 165, 250, 0.15); border-radius: 8px; background: rgba(15, 30, 52, 0.75);"></div>
    </div>

    <v-btn
      color="primary"
      variant="elevated"
      size="large"
      block
      :loading="loading"
      :disabled="loading || !isFormValid"
      @click="chargeBalance"
      prepend-icon="mdi-credit-card-plus"
    >
      Charge Balance
    </v-btn>
  </v-card>
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


