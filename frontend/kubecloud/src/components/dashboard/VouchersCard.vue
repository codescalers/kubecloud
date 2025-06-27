<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Vouchers & Credits</h3>
          <p class="dashboard-card-subtitle">Manage your promotional credits and vouchers</p>
        </div>
      </div>
    </div>
    <div class="vouchers-list">
      <div v-for="voucher in vouchers" :key="voucher.id" class="voucher-item list-item-interactive">
        <div class="voucher-content">
          <div class="voucher-info">
            <div class="voucher-icon">
              <v-icon size="24" color="primary">{{ voucher.icon }}</v-icon>
            </div>
            <div class="voucher-details">
              <h3 class="voucher-name">{{ voucher.name }}</h3>
              <p class="voucher-description">{{ voucher.description }}</p>
              <p class="voucher-expiry">Expires {{ voucher.expiryDate }}</p>
            </div>
          </div>
          <div class="voucher-value">
            <span class="voucher-amount">{{ voucher.amount }}</span>
            <v-btn variant="outlined" size="small" class="action-btn">Redeem</v-btn>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
interface Voucher {
  id: string | number
  name: string
  description: string
  amount: string
  expiryDate: string
  icon: string
  iconColor: string
}
const props = defineProps<{ vouchers: Voucher[] }>()
</script>

<style scoped>
.vouchers-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.voucher-item {
  padding: var(--space-8);
  border-radius: var(--radius-lg);
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid var(--color-border);
  transition: background 0.18s, border-color 0.18s;
}

.voucher-item:hover {
  background: rgba(30, 41, 59, 0.85);
  border-color: var(--color-border-light);
}

.voucher-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-12);
}

.voucher-info {
  display: flex;
  align-items: center;
  gap: var(--space-6);
  flex: 1;
  min-width: 0;
}

.voucher-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-md);
  background: rgba(59, 130, 246, 0.07);
  color: var(--color-primary);
  flex-shrink: 0;
}

.voucher-details {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  min-width: 0;
}

.voucher-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  font-size: var(--font-size-base);
  margin: 0;
  line-height: 1.2;
}

.voucher-description {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
  opacity: 0.9;
  margin: 0;
  line-height: 1.3;
}

.voucher-expiry {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin: 0;
  line-height: 1.2;
}

.voucher-value {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: var(--space-4);
  flex-shrink: 0;
}

.voucher-amount {
  font-weight: var(--font-weight-semibold);
  color: var(--color-success);
  font-size: var(--font-size-lg);
  line-height: 1;
}

.action-btn {
  background: transparent !important;
  border: 1px solid var(--color-border) !important;
  color: var(--color-text) !important;
  font-weight: var(--font-weight-medium);
  transition: background 0.18s, border-color 0.18s;
  white-space: nowrap;
  box-shadow: none !important;
}

.action-btn:hover {
  background: rgba(59, 130, 246, 0.07) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}

@media (max-width: 960px) {
  .voucher-content {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-6);
  }

  .voucher-value {
    align-self: stretch;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
}

@media (max-width: 600px) {
  .voucher-item {
    padding: var(--space-5);
  }

  .voucher-info {
    gap: var(--space-4);
  }

  .voucher-icon {
    width: 44px;
    height: 44px;
  }

  .voucher-details {
    gap: var(--space-2);
  }

  .voucher-value {
    gap: var(--space-2);
  }
}
</style>
