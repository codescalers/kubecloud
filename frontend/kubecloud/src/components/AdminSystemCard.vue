<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">System Stats</h2>
      <p class="section-subtitle">Platform health and performance metrics</p>
    </div>
    
    <div class="system-grid">
      <!-- Dependency Health -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-heart-pulse" color="#10B981" class="mr-2"></v-icon>
            Dependency Health
          </h3>
        </div>
        <div class="dependency-table">
          <div class="dependency-row header">
            <span class="dependency-name">Dependency</span>
            <span class="dependency-status">Status</span>
          </div>
          <div 
            v-for="(status, name) in metrics.dependency_health || {}" 
            :key="name"
            class="dependency-row"
          >
            <span class="dependency-name">{{ formatDependencyName(name) }}</span>
            <span 
              class="dependency-status"
              :class="status ? 'healthy' : 'unhealthy'"
            >
              {{ status ? 'Healthy' : 'Unhealthy' }}
            </span>
          </div>
          <div v-if="loading && (!metrics.dependency_health || Object.keys(metrics.dependency_health || {}).length === 0)" class="dependency-row">
            <span class="dependency-name">Loading...</span>
            <span class="dependency-status loading">Please wait</span>
          </div>
        </div>
      </div>

      <!-- System Info -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-information" color="#6366F1" class="mr-2"></v-icon>
            System Information
          </h3>
        </div>
        <div class="info-list">
          <div class="info-item">
            <span class="info-label">Version:</span>
            <span class="info-value">{{ metrics.system_info?.version || '1.0.0' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Uptime:</span>
            <span class="info-value">{{ formatUptime(metrics.system_info?.uptime) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Active Clusters:</span>
            <span class="info-value">{{ metrics.active_clusters || 0 }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Total Users:</span>
            <span class="info-value">{{ metrics.users_registered_total || 0 }}</span>
          </div>
        </div>
      </div>

      <!-- Stripe Payments -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-credit-card" color="#22C55E" class="mr-2"></v-icon>
            Stripe Payments
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value success">{{ metrics.stripe_payments_total?.success || 0 }}</div>
            <div class="metric-label">Success</div>
          </div>
          <div class="metric-item">
            <div class="metric-value error">{{ metrics.stripe_payments_total?.failure || 0 }}</div>
            <div class="metric-label">Failure</div>
          </div>
        </div>
      </div>

      <!-- Database Connections -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-database" color="#8B5CF6" class="mr-2"></v-icon>
            Database Connections
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ metrics.db_connections?.open || 0 }}</div>
            <div class="metric-label">Open</div>
          </div>
          <div class="metric-item">
            <div class="metric-value">{{ metrics.db_connections?.idle || 0 }}</div>
            <div class="metric-label">Idle</div>
          </div>
        </div>
      </div>

      <!-- HTTP Requests -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-web" color="#3B82F6" class="mr-2"></v-icon>
            HTTP Requests
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ metrics.http_requests?.total || 0 }}</div>
            <div class="metric-label">Total</div>
          </div>
          <div class="metric-item">
            <div class="metric-value success">{{ metrics.http_requests?.success || 0 }}</div>
            <div class="metric-label">Success</div>
          </div>
          <div class="metric-item">
            <div class="metric-value error">{{ metrics.http_requests?.error || 0 }}</div>
            <div class="metric-label">Errors</div>
          </div>
        </div>
      </div>

      <!-- Cluster Deployments -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-rocket-launch" color="#F59E0B" class="mr-2"></v-icon>
            Cluster Deployments
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value success">{{ metrics.cluster_deployments_total?.success || 0 }}</div>
            <div class="metric-label">Success</div>
          </div>
          <div class="metric-item">
            <div class="metric-value error">{{ metrics.cluster_deployments_total?.failure || 0 }}</div>
            <div class="metric-label">Failure</div>
          </div>
        </div>
      </div>

      <!-- Memory Usage -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-memory" color="#EC4899" class="mr-2"></v-icon>
            Memory Usage
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ Math.round((metrics.memory_usage?.used || 0) / 1024 / 1024) }}GB</div>
            <div class="metric-label">Used</div>
          </div>
          <div class="metric-item">
            <div class="metric-value">{{ Math.round((metrics.memory_usage?.total || 0) / 1024 / 1024) }}GB</div>
            <div class="metric-label">Total</div>
          </div>
        </div>
      </div>

      <!-- CPU Usage -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-cpu-64-bit" color="#06B6D4" class="mr-2"></v-icon>
            CPU Usage
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ metrics.cpu_usage?.percentage || 0 }}%</div>
            <div class="metric-label">Current</div>
          </div>
          <div class="metric-item">
            <div class="metric-value">{{ metrics.cpu_usage?.cores || 0 }}</div>
            <div class="metric-label">Cores</div>
          </div>
        </div>
      </div>

      <!-- Network Traffic -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-network" color="#8B5CF6" class="mr-2"></v-icon>
            Network Traffic
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ Math.round((metrics.network_traffic?.in || 0) / 1024 / 1024) }}MB</div>
            <div class="metric-label">In</div>
          </div>
          <div class="metric-item">
            <div class="metric-value">{{ Math.round((metrics.network_traffic?.out || 0) / 1024 / 1024) }}MB</div>
            <div class="metric-label">Out</div>
          </div>
        </div>
      </div>

      <!-- Disk Usage -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-harddisk" color="#10B981" class="mr-2"></v-icon>
            Disk Usage
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ Math.round((metrics.disk_usage?.used || 0) / 1024 / 1024 / 1024) }}GB</div>
            <div class="metric-label">Used</div>
          </div>
          <div class="metric-item">
            <div class="metric-value">{{ Math.round((metrics.disk_usage?.total || 0) / 1024 / 1024 / 1024) }}GB</div>
            <div class="metric-label">Total</div>
          </div>
        </div>
      </div>

      <!-- Active Sessions -->
      <div class="system-card">
        <div class="card-header">
          <h3 class="card-title">
            <v-icon icon="mdi-account-group" color="#F97316" class="mr-2"></v-icon>
            Active Sessions
          </h3>
        </div>
        <div class="metric-grid">
          <div class="metric-item">
            <div class="metric-value">{{ metrics.active_sessions?.current || 0 }}</div>
            <div class="metric-label">Current</div>
          </div>
          <div class="metric-item">
            <div class="metric-value">{{ metrics.active_sessions?.peak || 0 }}</div>
            <div class="metric-label">Peak</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'

interface Metrics {
  active_clusters?: number
  users_registered_total?: number
  db_connections?: {
    open: number
    idle: number
  }
  stripe_payments_total?: {
    success: number
    failure: number
  }
  http_requests?: {
    total: number
    success: number
    error: number
  }
  dependency_health?: Record<string, boolean>
  cluster_deployments_total?: {
    success: number
    failure: number
  }
  system_info?: {
    uptime: string
    version: string
  }
  memory_usage?: {
    used: number
    total: number
  }
  cpu_usage?: {
    percentage: number
    cores: number
  }
  network_traffic?: {
    in: number
    out: number
  }
  disk_usage?: {
    used: number
    total: number
  }
  active_sessions?: {
    current: number
    peak: number
  }
}

const props = defineProps<{
  metrics: Metrics
  loading?: boolean
}>()

function formatDependencyName(name: string): string {
  return name.split('_').map(word => 
    word.charAt(0).toUpperCase() + word.slice(1)
  ).join(' ')
}

function formatUptime(uptime: string | undefined): string {
  if (!uptime) return '0h 0m'
  // Parse Go duration format like "8760h0m0s" to "8760h 0m"
  const match = uptime.match(/(\d+)h(\d+)m/)
  if (match) {
    return `${match[1]}h ${match[2]}m`
  }
  return uptime
}
</script>

<style scoped>
.admin-section {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.section-header {
  text-align: left;
  margin-bottom: 1rem;
}

.dashboard-title {
  font-size: var(--font-size-3xl, 1.875rem);
  font-weight: var(--font-weight-bold, 700);
  margin-bottom: 0.5rem;
  line-height: 1.2;
  color: var(--color-text, #F8FAFC);
  letter-spacing: -0.5px;
}

.section-subtitle {
  font-size: var(--font-size-lg, 1.125rem);
  color: var(--color-text-secondary, #CBD5E1);
  line-height: 1.5;
  margin: 0;
  font-weight: var(--font-weight-normal, 400);
}

.system-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 8px;
  padding: 8px;
  align-items: start;
}

.system-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 12px;
  transition: all 0.2s ease;
}

.system-card:hover {
  border-color: var(--color-border-light, #475569);
  background: rgba(15, 30, 52, 0.75);
  transform: translateY(-1px);
}

/* Dependency Health - spans 4 columns, 2 rows (reduced width) */
.system-card:nth-child(1) {
  grid-column: span 4;
  grid-row: span 2;
}

/* System Info - spans 4 columns, 1 row */
.system-card:nth-child(5) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Stripe Payments - spans 4 columns, 1 row */
.system-card:nth-child(4) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Database Connections - spans 4 columns, 1 row */
.system-card:nth-child(2) {
  grid-column: span 4;
  grid-row: span 1;
}

/* HTTP Requests - spans 4 columns, 1 row */
.system-card:nth-child(3) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Cluster Deployments - spans 4 columns, 1 row */
.system-card:nth-child(6) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Memory Usage - spans 4 columns, 1 row */
.system-card:nth-child(7) {
  grid-column: span 4;
  grid-row: span 1;
}

/* CPU Usage - spans 4 columns, 1 row */
.system-card:nth-child(8) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Network Traffic - spans 4 columns, 1 row */
.system-card:nth-child(9) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Disk Usage - spans 4 columns, 1 row */
.system-card:nth-child(10) {
  grid-column: span 4;
  grid-row: span 1;
}

/* Active Sessions - spans 4 columns, 1 row */
.system-card:nth-child(11) {
  grid-column: span 4;
  grid-row: span 1;
}

.card-header {
  margin-bottom: 8px;
}

.card-title {
  font-size: var(--font-size-lg, 1.125rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text, #F8FAFC);
  margin: 0;
  display: flex;
  align-items: center;
}

.dependency-table {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background: rgba(15, 23, 42, 0.3);
  border-radius: var(--radius-lg, 0.5rem);
  padding: 0.75rem;
  border: 1px solid rgba(51, 65, 85, 0.2);
}

.dependency-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  border-radius: var(--radius-md, 0.375rem);
  transition: background-color var(--transition-normal, 0.2s);
}

.dependency-row:hover {
  background: rgba(51, 65, 85, 0.1);
}

.dependency-row.header {
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text-secondary, #CBD5E1);
  background: rgba(51, 65, 85, 0.2);
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
  margin-bottom: 0.25rem;
}

.dependency-row:last-child {
  border-bottom: none;
}

.dependency-name {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-text, #F8FAFC);
  font-weight: var(--font-weight-medium, 500);
}

.dependency-status {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: var(--font-weight-semibold, 600);
  padding: 0.375rem 0.75rem;
  border-radius: var(--radius-full, 9999px);
  text-transform: uppercase;
  letter-spacing: 0.025em;
  min-width: 80px;
  text-align: center;
}

.dependency-status.healthy {
  background: rgba(16, 185, 129, 0.15);
  color: #10B981;
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: 0 0 0 1px rgba(16, 185, 129, 0.1);
}

.dependency-status.unhealthy {
  background: rgba(239, 68, 68, 0.15);
  color: #EF4444;
  border: 1px solid rgba(239, 68, 68, 0.4);
  box-shadow: 0 0 0 1px rgba(239, 68, 68, 0.1);
}

.dependency-status.loading {
  background: rgba(59, 130, 246, 0.15);
  color: #3B82F6;
  border: 1px solid rgba(59, 130, 246, 0.4);
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.1);
}

.metric-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 8px;
}

.metric-item {
  text-align: center;
}

.metric-value {
  font-size: var(--font-size-2xl, 1.5rem);
  font-weight: var(--font-weight-bold, 700);
  color: var(--color-text, #F8FAFC);
  line-height: 1.2;
}

.metric-value.success {
  color: #10B981;
}

.metric-value.error {
  color: #EF4444;
}

.metric-label {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-text-secondary, #CBD5E1);
  margin-top: 0.25rem;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-text-secondary, #CBD5E1);
}

.info-value {
  font-size: var(--font-size-sm, 0.875rem);
  font-weight: var(--font-weight-medium, 500);
  color: var(--color-text, #F8FAFC);
}

@media (max-width: 1200px) {
  .system-grid {
    grid-template-columns: repeat(8, 1fr);
  }
  
  .system-card:nth-child(1) {
    grid-column: span 8;
    grid-row: span 1;
  }
  
  .system-card:nth-child(5),
  .system-card:nth-child(4) {
    grid-column: span 4;
    grid-row: span 1;
  }
  
  .system-card:nth-child(2),
  .system-card:nth-child(3),
  .system-card:nth-child(6) {
    grid-column: span 4;
    grid-row: span 1;
  }

  .system-card:nth-child(7),
  .system-card:nth-child(8),
  .system-card:nth-child(9),
  .system-card:nth-child(10),
  .system-card:nth-child(11) {
    grid-column: span 4;
    grid-row: span 1;
  }
}

@media (max-width: 768px) {
  .system-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .system-card:nth-child(1),
  .system-card:nth-child(2),
  .system-card:nth-child(3),
  .system-card:nth-child(4),
  .system-card:nth-child(5),
  .system-card:nth-child(6),
  .system-card:nth-child(7),
  .system-card:nth-child(8),
  .system-card:nth-child(9),
  .system-card:nth-child(10),
  .system-card:nth-child(11) {
    grid-column: span 1;
    grid-row: span 1;
  }
  
  .metric-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style> 