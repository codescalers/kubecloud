<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Manual Credit</h3>
      <p class="section-subtitle">Apply credits to user accounts</p>
    </div>
    <v-form @submit.prevent="$emit('applyManualCredit')" class="credit-form">
      <div class="form-row">
        <v-text-field
          v-model.number="creditAmountLocal"
          label="Amount ($)"
          type="number"
          prepend-inner-icon="mdi-currency-usd"
          variant="outlined"
          min="0.01"
          step="0.01"
          density="comfortable"
          required
          class="form-field"
        />
        <v-text-field
          v-model="creditReasonLocal"
          label="Reason/Memo"
          prepend-inner-icon="mdi-note-text"
          variant="outlined"
          density="comfortable"
          min-length="3"
          required
          class="form-field"
        />
      </div>
      <v-btn type="submit" color="primary" variant="elevated" class="btn-primary">
        <v-icon icon="mdi-cash-plus" class="mr-2"></v-icon>
        Apply Credit
      </v-btn>
    </v-form>
    <v-alert v-if="creditResult" type="success" variant="tonal" class="mt-4">{{ creditResult }}</v-alert>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
const props = defineProps({
  creditAmount: Number,
  creditReason: String,
  creditResult: String
})
const emit = defineEmits(['applyManualCredit', 'update:creditAmount', 'update:creditReason'])
const creditAmountLocal = ref(props.creditAmount)
const creditReasonLocal = ref(props.creditReason)
watch(() => props.creditAmount, val => { creditAmountLocal.value = val })
watch(() => props.creditReason, val => { creditReasonLocal.value = val })
watch(creditAmountLocal, val => emit('update:creditAmount', val))
watch(creditReasonLocal, val => emit('update:creditReason', val))
</script>
