<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Payments</h3>
          <p class="dashboard-card-subtitle">View your payment records</p>
        </div>
      </div>
    </div>
    <PendingRecordsTable :pendingRecords="pendingRecords" :loading="loading" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { userService, type PendingRecord } from '../../utils/userService'
import PendingRecordsTable from './PendingRecordsTable.vue'

const pendingRecords = ref<PendingRecord[]>([])
const loading = ref(true)
onMounted(async () => {
  try {
    loading.value = true
    const response = await userService.listUserPendingRecords()
    pendingRecords.value = response || []
    loading.value = false
  } catch (error) {
    console.error('Failed to load user payments:', error)
    loading.value = false
  }
})
</script>
