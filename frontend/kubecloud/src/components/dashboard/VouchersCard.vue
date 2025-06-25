<template>
  <div class="dashboard-card vouchers-card">
    <h2 class="dashboard-card-title mb-6">Vouchers & Credits</h2>
    <div class="vouchers-list mb-6">
      <div v-for="voucher in vouchers" :key="voucher.id" class="voucher-item">
        <div class="voucher-content">
          <div class="voucher-info">
            <div class="voucher-icon">
              <v-icon size="32" :color="'accent'">{{ voucher.icon }}</v-icon>
            </div>
            <div class="voucher-details">
              <h3 class="voucher-name">{{ voucher.name }}</h3>
              <p class="voucher-description text-muted">{{ voucher.description }}</p>
              <p class="voucher-expiry text-muted">Expires {{ voucher.expiryDate }}</p>
            </div>
          </div>
          <div class="voucher-value">
            <span class="voucher-amount">{{ voucher.amount }}</span>
            <v-btn class="action-btn">Redeem</v-btn>
          </div>
        </div>
      </div>
    </div>
    <v-btn class="action-btn add-voucher-btn">
      <v-icon icon="mdi-plus" size="24" class="mr-2"></v-icon>
      Add Voucher Code
    </v-btn>
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
.dashboard-card.vouchers-card {
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  color: var(--color-text);
  padding: 1.25rem;
  min-width: 100%;
}

.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}

.vouchers-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.voucher-item {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--rounded);
  padding: 1.2rem;
  transition: border-color var(--transition), background var(--transition);
  position: relative;
  overflow: hidden;
}

.voucher-item:hover {
  border-color: var(--color-accent);
}

.voucher-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.voucher-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  flex: 1;
}

.voucher-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--rounded);
  background: none;
}

.voucher-name {
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.1rem;
}

.voucher-description, .voucher-expiry {
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.voucher-value {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.voucher-amount {
  font-weight: 600;
  color: var(--color-accent);
  font-size: 1rem;
}

.action-btn {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}

.action-btn:hover {
  background: var(--color-accent-hover);
}

.add-voucher-btn {
  margin-top: 0.5rem;
  font-weight: 600;
  height: 40px;
}
</style>
