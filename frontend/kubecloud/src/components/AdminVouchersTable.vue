<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">Voucher Management</h2>
      <p class="section-subtitle">Generate and manage platform vouchers</p>
    </div>
    <div class="dashboard-card">
      <div class="dashboard-card-header">
        <h3 class="dashboard-card-title">Generate Vouchers</h3>
        <p class="dashboard-card-subtitle">Create new vouchers for user promotions</p>
      </div>
      <v-form @submit.prevent="$emit('generateVouchers')" class="voucher-form">
        <div class="form-row">
          <v-text-field 
            v-model.number="voucherValueLocal" 
            label="Voucher Value ($)" 
            type="number" 
            prepend-inner-icon="mdi-currency-usd" 
            variant="outlined" 
            min="1" 
            density="comfortable"
            required 
            class="form-field"
          />
          <v-text-field 
            v-model.number="voucherCountLocal" 
            label="Number of Vouchers" 
            type="number" 
            prepend-inner-icon="mdi-pound" 
            variant="outlined" 
            min="1" 
            density="comfortable"
            required 
            class="form-field"
          />
          <v-text-field 
            v-model.number="voucherExpiryLocal" 
            label="Expiry (days)" 
            type="number" 
            prepend-inner-icon="mdi-calendar-clock" 
            variant="outlined" 
            min="1" 
            density="comfortable"
            required 
            class="form-field"
          />
        </div>
        <v-btn type="submit" color="primary" variant="elevated" class="btn-primary">
          <v-icon icon="mdi-ticket-percent" class="mr-2"></v-icon>
          Generate
        </v-btn>
      </v-form>
      <v-alert v-if="voucherResult" type="success" variant="tonal" class="mt-4">{{ voucherResult }}</v-alert>
    </div>
    <div class="dashboard-card">
      <div class="dashboard-card-header">
        <h3 class="dashboard-card-title">All Vouchers</h3>
        <p class="dashboard-card-subtitle">List of all generated vouchers</p>
      </div>
      <div class="table-container">
        <v-data-table
        :headers="[
            { title: 'Voucher', key: 'code' },
            { title: 'Value', key: 'value' },
            { title: 'Redeemed', key: 'redeemed' },
            { title: 'Created At', key: 'created_at' },
            { title: 'Expires At', key: 'expires_at' }
          ]"
          :items="paginatedVouchers"
          class="admin-table"
          hide-default-footer
          density="comfortable"
        >
        <template #item.value="{ item }">
            {{ item.value }}
          </template>
          <template #item.redeemed="{ item }">
            {{ item.redeemed === true ? 'Yes' : 'No' }}
          </template>
          <template #item.created_at="{ item }">
            <span>{{ formatDate(item.created_at) }}</span>
          </template>
          <template #item.expires_at="{ item }">
            <span>{{ formatDate(item.expires_at) }}</span>
          </template>
        </v-data-table>
        <div class="pagination-container mt-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            color="primary"
            circle
            size="small"
          />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, watch, computed } from 'vue'
const props = defineProps({
  voucherValue: Number,
  voucherCount: Number,
  voucherExpiry: Number,
  voucherResult: String,
  vouchers: {
    type: Array as () => any[],
    default: () => []
  }
})
const emit = defineEmits(['generateVouchers', 'update:voucherValue', 'update:voucherCount', 'update:voucherExpiry'])
const voucherValueLocal = ref(props.voucherValue)
const voucherCountLocal = ref(props.voucherCount)
const voucherExpiryLocal = ref(props.voucherExpiry)
watch(() => props.voucherValue, val => { voucherValueLocal.value = val })
watch(() => props.voucherCount, val => { voucherCountLocal.value = val })
watch(() => props.voucherExpiry, val => { voucherExpiryLocal.value = val })
watch(voucherValueLocal, val => emit('update:voucherValue', val))
watch(voucherCountLocal, val => emit('update:voucherCount', val))
watch(voucherExpiryLocal, val => emit('update:voucherExpiry', val))

// Pagination state
const currentPage = ref(1)
const pageSize = 10
const totalPages = computed(() => Math.ceil((props.vouchers?.length || 0) / pageSize))
const paginatedVouchers = computed<any[]>(() => {
  const start = (currentPage.value - 1) * pageSize
  return props.vouchers?.slice(start, start + pageSize) || []
})

function formatDate(dateStr: string) {
  if (!dateStr) return '-';
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return d.toLocaleString(undefined, { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script> 

export default {}; 