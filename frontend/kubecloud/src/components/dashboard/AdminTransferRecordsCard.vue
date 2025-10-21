<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Transactions</h3>
          <p class="dashboard-card-subtitle">View transactions records</p>
        </div>
      </div>
    </div>
    <TransferRecordsTable
      :transferRecords="transferRecords"
      :loading="loading"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { type TransferRecord } from '../../utils/adminService'
import TransferRecordsTable from './TransferRecordsTable.vue'
import { useNotificationStore } from '../../stores/notifications'
import { adminService } from '../../utils/adminService'

const transferRecords = ref<TransferRecord[]>([])
const notificationStore = useNotificationStore()

onMounted(async () => {
  await loadTransferRecords()
})

const loading = ref(false)

async function loadTransferRecords() {
  loading.value = true
  try {
    const response = await adminService.listTransferRecords()
    transferRecords.value = response || []
  } catch (error) {
    console.error('Failed to load transfer records:', error)
    notificationStore.error('Error', 'Failed to load transfer records')
  } finally {
    loading.value = false
  }
}


</script>

<style scoped>
/* Card styling is inherited from global styles */
</style>
