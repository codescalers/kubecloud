<template>
  <v-card color="surface-variant" class="pa-6">
    <div class="mb-6">
      <h3 class="text-h5 font-weight-semibold mb-2">Manual Credit</h3>
      <p class="text-body-1 text-medium-emphasis">Apply credits to user accounts</p>
    </div>

    <v-form @submit.prevent="$emit('applyManualCredit')" class="d-flex flex-column ga-4">
      <div class="d-flex flex-column flex-sm-row ga-4">
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
          class="flex-grow-1"
        />
        <v-text-field
          v-model="creditReasonLocal"
          label="Reason/Memo"
          prepend-inner-icon="mdi-note-text"
          variant="outlined"
          density="comfortable"
          required
          class="flex-grow-1"
        />
      </div>
      <v-btn type="submit" color="primary" variant="elevated">
        <v-icon icon="mdi-cash-plus" class="mr-2"></v-icon>
        Apply Credit
      </v-btn>
    </v-form>
    <v-alert v-if="creditResult" type="success" variant="tonal" class="mt-4">{{ creditResult }}</v-alert>
  </v-card>
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
