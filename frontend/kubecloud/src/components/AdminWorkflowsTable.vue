<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">Workflows</h2>
      <p class="section-subtitle">All platform workflows</p>
    </div>
    <div class="dashboard-card">
      <div class="dashboard-card-header">
        <div class="header-row">
          <div>
            <h3 class="dashboard-card-title">All Workflows</h3>
            <p class="dashboard-card-subtitle">List of all workflow executions</p>
          </div>
          <div class="filter-controls">
            <v-select
              v-model="statusFilter"
              :items="statusOptions"
              label="Filter by Status"
              variant="outlined"
              density="compact"
              clearable
              class="status-filter"
              @update:modelValue="filterByStatus"
            />
            <v-btn 
              variant="outlined" 
              size="small" 
              class="refresh-btn"
              @click="refreshWorkflows"
              :loading="loading"
            >
              <v-icon icon="mdi-refresh" size="18" class="mr-1"></v-icon>
              Refresh
            </v-btn>
          </div>
        </div>
      </div>
      <div class="table-container">
        <v-data-table
          :headers="headers"
          :items="workflows"
          class="admin-table"
          density="comfortable"
          :loading="loading"
          :items-length="totalWorkflows"
          v-model:page="page"
          v-model:items-per-page="itemsPerPage"
          @update:page="loadWorkflows"
          @update:items-per-page="loadWorkflows"
        >
          <template v-slot:item.display_name="{ item }">
            <div class="workflow-name">
              <span class="name-text">{{ item.display_name || item.name }}</span>
              <span class="name-template text-caption">({{ item.name }})</span>
            </div>
          </template>
          <template v-slot:item.status="{ item }">
            <v-chip 
              :color="getStatusColor(item.status)" 
              size="small"
              variant="flat"
            >
              {{ item.status }}
            </v-chip>
          </template>
          <template v-slot:item.progress="{ item }">
            <div class="progress-cell">
              <span>{{ Math.min(item.current_step + 1, item.total_steps) }} / {{ item.total_steps }}</span>
              <v-progress-linear
                :model-value="(Math.min(item.current_step + 1, item.total_steps) / item.total_steps) * 100"
                :color="getStatusColor(item.status)"
                height="4"
                rounded
                class="progress-bar"
              />
            </div>
          </template>
          <template v-slot:item.step_name="{ item }">
            <span class="step-name">{{ item.step_name || '-' }}</span>
          </template>
          <template v-slot:item.user_id="{ item }">
            <span>{{ item.user_id || '-' }}</span>
          </template>
          <template v-slot:item.created_at="{ item }">
            <span>{{ formatDate(item.created_at) }}</span>
          </template>
          <template v-slot:item.actions="{ item }">
            <v-btn size="small" variant="outlined" class="action-btn" @click="viewWorkflow(item)">
              <v-icon icon="mdi-eye" size="16" class="mr-1"></v-icon>
              View
            </v-btn>
          </template>
        </v-data-table>
      </div>
    </div>
    <v-dialog v-model="showWorkflowModal" max-width="700" persistent>
      <v-card v-if="selectedWorkflow" class="pa-4" style="background: rgba(16,24,39,0.98); border-radius: 18px;">
        <v-card-title class="text-h6 font-weight-bold mb-2">Workflow Details</v-card-title>
        <v-card-text>
          <div class="workflow-details">
            <div class="detail-row">
              <strong>UUID:</strong>
              <span class="monospace">{{ selectedWorkflow.uuid }}</span>
            </div>
            <div class="detail-row">
              <strong>Display Name:</strong>
              <span>{{ selectedWorkflow.display_name || selectedWorkflow.name }}</span>
            </div>
            <div class="detail-row">
              <strong>Template Name:</strong>
              <span>{{ selectedWorkflow.name }}</span>
            </div>
            <div class="detail-row">
              <strong>Status:</strong>
              <v-chip :color="getStatusColor(selectedWorkflow.status)" size="small" variant="flat">
                {{ selectedWorkflow.status }}
              </v-chip>
            </div>
            <div class="detail-row">
              <strong>Progress:</strong>
              <span>Step {{ Math.min(selectedWorkflow.current_step + 1, selectedWorkflow.total_steps) }} of {{ selectedWorkflow.total_steps }}{{ selectedWorkflow.step_name && selectedWorkflow.step_name !== '-' ? ` (${selectedWorkflow.step_name})` : '' }}</span>
            </div>
            <div class="detail-row">
              <strong>User ID:</strong>
              <span>{{ selectedWorkflow.user_id || 'N/A' }}</span>
            </div>
            <div class="detail-row">
              <strong>Queue:</strong>
              <span>{{ selectedWorkflow.queue_name || 'default' }}</span>
            </div>
            <div class="detail-row">
              <strong>Created At:</strong>
              <span>{{ formatDate(selectedWorkflow.created_at) }}</span>
            </div>
            <div class="detail-section" v-if="selectedWorkflow.metadata && Object.keys(selectedWorkflow.metadata).length > 0">
              <strong>Metadata:</strong>
              <pre class="json-display">{{ JSON.stringify(selectedWorkflow.metadata, null, 2) }}</pre>
            </div>
            <div class="detail-section">
              <strong>State:</strong>
              <pre class="json-display">{{ formatState(selectedWorkflow.state) }}</pre>
            </div>
          </div>
        </v-card-text>
        <v-card-actions class="justify-end mt-2">
          <v-btn text color="grey-lighten-1" @click="closeWorkflowModal">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminService, type AdminWorkflow } from '../utils/adminService'
import { formatDate } from '../utils/uiUtils'

const page = ref(1)
const itemsPerPage = ref(10)
const loading = ref(false)
const workflows = ref<AdminWorkflow[]>([])
const statusFilter = ref<string | null>(null)
const totalWorkflows = ref(0)
const totalPages = ref(0)

const statusOptions = [
  { title: 'Pending', value: 'pending' },
  { title: 'Running', value: 'running' },
  { title: 'Completed', value: 'completed' },
  { title: 'Failed', value: 'failed' },
]

const headers = [
  { title: 'Display Name', key: 'display_name', width: '200px' },
  { title: 'Status', key: 'status', width: '120px' },
  { title: 'Progress', key: 'progress', width: '150px' },
  { title: 'Current Step', key: 'step_name', width: '150px' },
  { title: 'User ID', key: 'user_id', width: '100px' },
  { title: 'Created At', key: 'created_at', width: '180px' },
  { title: 'Actions', key: 'actions', sortable: false, width: '100px' }
]

const showWorkflowModal = ref(false)
const selectedWorkflow = ref<AdminWorkflow | null>(null)

function getStatusColor(status: string): string {
  switch (status) {
    case 'pending':
      return 'warning'
    case 'running':
      return 'info'
    case 'completed':
      return 'success'
    case 'failed':
      return 'error'
    default:
      return 'grey'
  }
}

function viewWorkflow(workflow: AdminWorkflow) {
  selectedWorkflow.value = workflow
  showWorkflowModal.value = true
}

function closeWorkflowModal() {
  showWorkflowModal.value = false
  selectedWorkflow.value = null
}

function formatState(state: Record<string, any>): string {
  if (!state) return '{}'
  // Filter out sensitive data
  const sensitiveKeys = ['mnemonic', 'password', 'secret', 'token', 'key', 'api_key', 'private_key', 'access_token', 'refresh_token', 'auth', 'credential']
  const filtered = { ...state }
  
  for (const key of Object.keys(filtered)) {
    const lowerKey = key.toLowerCase()
    if (sensitiveKeys.some(sensitiveKey => lowerKey.includes(sensitiveKey))) {
      filtered[key] = '[REDACTED]'
    }
  }
  
  return JSON.stringify(filtered, null, 2)
}

async function loadWorkflows() {
  loading.value = true
  try {
    const data = await adminService.listWorkflowsPaginated(statusFilter.value || undefined, page.value, itemsPerPage.value)
    // Ensure we have valid data
    workflows.value = Array.isArray(data.workflows) ? data.workflows : []
    totalWorkflows.value = typeof data.total === 'number' ? data.total : 0
    totalPages.value = typeof data.total_pages === 'number' ? data.total_pages : 0
  } catch (error) {
    console.error('Failed to load workflows:', error)
    // Reset to empty on error
    workflows.value = []
    totalWorkflows.value = 0
    totalPages.value = 0
  } finally {
    loading.value = false
  }
}

async function filterByStatus() {
  page.value = 1
  await loadWorkflows()
}

async function refreshWorkflows() {
  await loadWorkflows()
}

onMounted(async () => {
  await loadWorkflows()
})
</script>

<style scoped>
.admin-section {
  margin-bottom: 2rem;
}
.section-header {
  margin-bottom: 1.5rem;
}
.dashboard-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}
.section-subtitle {
  color: #94a3b8;
  font-size: 1rem;
}
.dashboard-card {
  background: rgba(10, 25, 47, 0.85);
  border: 1px solid var(--color-border, #334155);
  border-radius: 1rem;
  padding: 1.5rem;
  margin-bottom: 2rem;
}
.dashboard-card-header {
  margin-bottom: 1rem;
}
.header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
}
.filter-controls {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
.status-filter {
  min-width: 180px;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 500;
}
.dashboard-card-subtitle {
  color: #64748b;
  font-size: 0.95rem;
}
.table-container {
  margin-top: 1rem;
}
.workflow-name {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}
.name-text {
  font-weight: 500;
}
.name-template {
  color: #64748b;
  font-size: 0.75rem;
}
.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 100px;
}
.progress-bar {
  margin-top: 0.25rem;
}
.step-name {
  font-size: 0.875rem;
  color: #cbd5e1;
}
.action-btn {
  background: transparent !important;
  border: 1px solid var(--color-border) !important;
  color: var(--color-text) !important;
  font-weight: 500;
  transition: all 0.2s;
}
.action-btn:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}
.refresh-btn {
  background: transparent !important;
  border: 1px solid var(--color-border) !important;
  color: var(--color-text) !important;
  font-weight: 500;
  transition: all 0.2s;
}
.refresh-btn:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}
.workflow-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.detail-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.detail-row strong {
  min-width: 120px;
  color: #94a3b8;
}
.detail-section {
  margin-top: 0.5rem;
}
.detail-section strong {
  display: block;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}
.monospace {
  font-family: monospace;
  font-size: 0.875rem;
  background: rgba(0, 0, 0, 0.2);
  padding: 0.125rem 0.5rem;
  border-radius: 0.25rem;
}
.json-display {
  background: rgba(0, 0, 0, 0.3);
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  font-size: 0.8rem;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
