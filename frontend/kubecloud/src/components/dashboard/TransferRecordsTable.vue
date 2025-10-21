<template>
  <div class="records-table-container">
    <v-data-table :loading="loading" :headers="headers" :items="transferRecords" class="records-table"
      :items-per-page="5" :no-data-text="'No records found'" density="comfortable">
      <template v-slot:[`item.id`]="{ item }">
        <span>{{ item.id }}</span>
      </template>
      <template v-slot:[`item.user_id`]="{ item }">
        <span>{{ item.user_id }}</span>
      </template>
      <template  v-slot:[`item.username`]="{ item }">
        <span>{{ item.username }}</span>
      </template>
      <template v-slot:[`item.operation`]="{ item }">
        <span>{{ capitalize(item.operation) }}</span>
      </template>
      <template v-slot:[`item.state`]="{ item }">
        <v-chip :color="getStatusColor(item)" size="small" class="status-chip">
          {{ capitalize(item.state) }}
        </v-chip>
      </template>
      <template v-slot:[`item.tft_amount_in_whole_unit`]="{ item }">
        <span>{{ +item.tft_amount_in_whole_unit.toFixed(2) }}</span>
      </template>
      <template v-slot:[`item.created_at`]="{ item }">
        <span>{{ formatDate(item.created_at) }}</span>
      </template>
      <template v-slot:[`item.updated_at`]="{ item }">
        <span>{{ formatDate(item.updated_at) }}</span>
      </template>


    </v-data-table>
  </div>
</template>

<script setup lang="ts">
import { type RecordState, type TransferRecord } from '../../utils/adminService'
import { formatDate } from '../../utils/uiUtils'
import { capitalize } from 'vue'
const props = defineProps({
  transferRecords: {
    type: Array as () => TransferRecord[],
    required: true,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const headers = [
  { title: 'ID', key: 'id' },
  { title: 'User ID', key: 'user_id' },
  { title: 'Username', key: 'username' },
  { title: 'Record Date', key: 'created_at' },
  { title: 'Updated Date', key: 'updated_at' },
  { title: 'Operation', key: 'operation' },
  { title: 'Status', key: 'state' },
  { title: 'TFT Amount', key: 'tft_amount_in_whole_unit' },
]



function getStatusColor(item: TransferRecord): string {
  if (item.state === 'success') {
    return 'success'
  }
  if (item.state === 'failed') {
    return 'error'
  }
  return 'gray'
}
</script>

<style scoped>
.records-table-container {
  margin-bottom: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.records-table {
  background: transparent;
  width: 100%;
}

.status-chip {
  font-size: 0.75rem;
  font-weight: var(--font-weight-medium);
}
</style>
