<template>
  <div class="admin-manual-credit">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Manual Credit</h3>
      <p class="section-subtitle">Apply credits to user accounts</p>
    </div>
    
    <v-form ref="formRef" @submit.prevent="handleSubmit" class="credit-form" v-model="isFormValid">
      <div class="form-row">
        <v-text-field
          v-model.number="creditAmount"
          label="Amount ($)"
          type="number"
          prepend-inner-icon="mdi-currency-usd"
          variant="outlined"
          density="comfortable"
          class="form-field mb-3"
          :rules="[RULES.creditAmount]"
          placeholder="Enter amount (e.g., 25.50)"
        />
        <v-text-field
          v-model="creditReason"
          label="Reason/Memo"
          prepend-inner-icon="mdi-note-text"
          variant="outlined"
          density="comfortable"
          class="form-field mb-3"
          :rules="[RULES.creditMemo]"
          placeholder="Enter reason for credit (e.g., Account adjustment)"
          counter="255"
        />
      </div>
      
      <v-btn
        type="submit"
        color="primary"
        variant="elevated"
        class="btn-primary"
        :disabled="isSubmitting || !isFormValid"
        :loading="isSubmitting"
      >
        <v-icon icon="mdi-cash-plus" class="mr-2"></v-icon>
        {{ isSubmitting ? 'Applying Credit...' : 'Apply Credit' }}
      </v-btn>
    </v-form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RULES } from '../utils/validation'
import { adminService } from '../utils/adminService'

interface Props {
  userId: number
  userEmail: string
}

const props = defineProps<Props>()
const emit = defineEmits(['creditApplied', 'close'])
const formRef = ref()
const isSubmitting = ref(false)
const isFormValid = ref(false)
const creditAmount = ref(0)
const creditReason = ref('')

const handleSubmit = async () => {
  try {
    isSubmitting.value = true    
    await adminService.creditUser(props.userId, {
      amount: creditAmount.value,
      memo: creditReason.value.trim()
    })
    
    // Emit success event to parent
    emit('creditApplied', {
      userId: props.userId,
      amount: creditAmount.value,
      memo: creditReason.value.trim()
    })
    
    setTimeout(() => {
      emit('close')
    }, 1500)
    
  } catch (error: any) {
    console.error('Credit application error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>
