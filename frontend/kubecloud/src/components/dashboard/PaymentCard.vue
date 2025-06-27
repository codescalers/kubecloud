<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Payment Methods</h3>
          <p class="dashboard-card-subtitle">Manage your payment methods and billing information</p>
        </div>
      </div>
    </div>
    <div class="payment-methods">
      <div v-for="method in paymentMethods" :key="method.id" class="payment-method-item list-item-interactive">
        <div class="payment-method-content">
          <div class="payment-method-info">
            <div class="payment-method-icon">
              <div v-if="method.name.toLowerCase().includes('paypal')" class="paypal-icon">
                <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
                  <path d="M20.067 8.478c.492.315.844.825.844 1.478 0 .653-.352 1.163-.844 1.478-.492.315-1.163.478-1.844.478H16.5v-2.956h1.723c.681 0 1.352.163 1.844.478zM7.933 8.478c.492-.315 1.163-.478 1.844-.478H11.5v2.956H9.777c-.681 0-1.352-.163-1.844-.478-.492-.315-.844-.825-.844-1.478 0-.653.352-1.163.844-1.478z"/>
                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.5 15.5v-2.956h1.723c.681 0 1.352-.163 1.844-.478.492-.315.844-.825.844-1.478 0-.653-.352-1.163-.844-1.478-.492-.315-1.163-.478-1.844-.478H10.5V8.5h3.723c1.362 0 2.704.326 3.688.956.984.63 1.688 1.65 1.688 2.956 0 1.306-.704 2.326-1.688 2.956-.984.63-2.326.956-3.688.956H10.5z"/>
                </svg>
              </div>
              <v-icon v-else size="24" color="primary">{{ method.icon }}</v-icon>
            </div>
            <div class="payment-method-details">
              <h3 class="payment-method-name">{{ method.name }}</h3>
              <p class="payment-method-number">{{ method.maskedNumber }}</p>
            </div>
          </div>
          <div class="payment-method-actions">
            <v-btn variant="outlined" size="small" class="action-btn">Edit</v-btn>
            <v-btn variant="outlined" size="small" class="action-btn">Remove</v-btn>
          </div>
        </div>
      </div>
    </div>
    <div class="add-payment-section">
      <v-btn variant="outlined" class="add-payment-btn action-btn">
        <v-icon icon="mdi-plus" size="20" class="mr-2"></v-icon>
        Add Payment Method
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
interface PaymentMethod {
  id: string | number
  name: string
  maskedNumber: string
  icon: string
  iconColor: string
}
const props = defineProps<{ paymentMethods: PaymentMethod[] }>()
</script>

<style scoped>
.payment-methods {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  margin-bottom: var(--space-10);
}

.payment-method-item {
  padding: var(--space-8);
  border-radius: var(--radius-lg);
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid var(--color-border);
  transition: background 0.18s, border-color 0.18s;
}

.payment-method-item:hover {
  background: rgba(30, 41, 59, 0.85);
  border-color: var(--color-border-light);
}

.payment-method-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-12);
}

.payment-method-info {
  display: flex;
  align-items: center;
  gap: var(--space-6);
  flex: 1;
  min-width: 0;
}

.payment-method-icon {
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

.payment-method-details {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  min-width: 0;
}

.payment-method-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  font-size: var(--font-size-base);
  margin: 0;
  line-height: 1.2;
}

.payment-method-number {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin: 0;
  line-height: 1.3;
}

.payment-method-actions {
  display: flex;
  gap: var(--space-4);
  flex-shrink: 0;
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

.add-payment-section {
  display: flex;
  justify-content: center;
  padding-top: var(--space-8);
  border-top: 1px solid var(--color-border);
}

.add-payment-btn {
  font-weight: var(--font-weight-medium);
  height: 44px;
  min-width: 220px;
}

@media (max-width: 960px) {
  .payment-method-content {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-6);
  }

  .payment-method-actions {
    align-self: stretch;
    justify-content: flex-end;
  }
}

@media (max-width: 600px) {
  .payment-method-item {
    padding: var(--space-5);
  }

  .payment-method-info {
    gap: var(--space-4);
  }

  .payment-method-icon {
    width: 44px;
    height: 44px;
  }

  .payment-method-details {
    gap: var(--space-2);
  }

  .payment-method-actions {
    gap: var(--space-2);
  }
  
  .add-payment-btn {
    min-width: 100%;
  }
}
</style>
