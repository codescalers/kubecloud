<template>
  <v-card class="pa-6" variant="outlined" rounded="xl">
    <v-card-title class="py-0 text-h5 font-weight-bold mb-1">Add Funds</v-card-title>
    <v-card-subtitle class="text-body-1 dashboard-card-subtitle">Add funds to your account balance</v-card-subtitle>
    <v-divider class="my-4"></v-divider>

    <v-row class="mb-2">
      <v-col cols="12" class="d-flex justify-space-between align-center">
        <div class="text-body-1">Current Balance:</div>
        <div class="text-body-1 text-success font-weight-bold">${{ userStore.netBalance.toFixed(2) }}</div>
      </v-col>
      <v-col v-if="userStore.pendingBalance > 0" cols="12" class="d-flex justify-space-between align-center">
        <div class="text-body-1">Pending Balance:</div>
        <div class="text-body-1 text-medium-emphasis font-weight-medium">${{ userStore.pendingBalance.toFixed(2) }}</div>
      </v-col>
    </v-row>

    <v-row class="mb-4" align="center">
      <v-col cols="12" class="d-flex align-center flex-wrap">
        <div class="me-4">Amount:</div>
        <div class="d-flex flex-wrap ga-2">
          <v-btn
            v-for="preset in presets"
            :key="preset"
            :variant="typeof amount === 'number' && amount === preset ? 'flat' : 'tonal'"
            :color="typeof amount === 'number' && amount === preset ? 'primary' : undefined"
            size="large"
            @click="selectAmount(preset)"
          >
            {{ preset }}
          </v-btn>
          <v-btn
            :variant="typeof amount === 'string' && amount === 'custom' ? 'flat' : 'tonal'"
            :color="typeof amount === 'string' && amount === 'custom' ? 'primary' : undefined"
            size="large"
            @click="selectAmount('custom')"
          >
            Custom
          </v-btn>
          <v-text-field
            v-if="typeof amount === 'string' && amount === 'custom'"
            v-model.number="customAmount"
            type="number"
            min="1"
            density="comfortable"
            variant="outlined"
            label="Custom amount"
            hide-details="auto"
            class="ms-2"
            @focus="selectAmount('custom')"
            style="max-width: 140px"
          />
        </div>
      </v-col>
    </v-row>

    <v-sheet variant="outlined" rounded="lg" class="pa-3 mb-4">
      <div id="stripe-card-element"></div>
    </v-sheet>

    <v-btn
      color="primary"
      :loading="loading"
      :disabled="loading || !isFormValid"
      @click="chargeBalance"
      prepend-icon="mdi-credit-card-plus"
      size="large"
      rounded="lg"
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

