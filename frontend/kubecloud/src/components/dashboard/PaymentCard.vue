<template>
  <div class="dashboard-card payment-card">
    <h2 class="dashboard-card-title mb-6">Payment Methods</h2>
    <div class="payment-methods mb-6">
      <div v-for="method in paymentMethods" :key="method.id" class="payment-method-item">
        <div class="payment-method-content">
          <div class="payment-method-info">
            <div class="payment-method-icon">
              <div v-if="method.name.toLowerCase().includes('paypal')" class="paypal-icon">
                <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
                  <path d="M20.067 8.478c.492.315.844.825.844 1.478 0 .653-.352 1.163-.844 1.478-.492.315-1.163.478-1.844.478H16.5v-2.956h1.723c.681 0 1.352.163 1.844.478zM7.933 8.478c.492-.315 1.163-.478 1.844-.478H11.5v2.956H9.777c-.681 0-1.352-.163-1.844-.478-.492-.315-.844-.825-.844-1.478 0-.653.352-1.163.844-1.478z"/>
                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.5 15.5v-2.956h1.723c.681 0 1.352-.163 1.844-.478.492-.315.844-.825.844-1.478 0-.653-.352-1.163-.844-1.478-.492-.315-1.163-.478-1.844-.478H10.5V8.5h3.723c1.362 0 2.704.326 3.688.956.984.63 1.688 1.65 1.688 2.956 0 1.306-.704 2.326-1.688 2.956-.984.63-2.326.956-3.688.956H10.5z"/>
                </svg>
              </div>
              <v-icon v-else size="32" :color="'accent'">{{ method.icon }}</v-icon>
            </div>
            <div class="payment-method-details">
              <h3 class="payment-method-name">{{ method.name }}</h3>
              <p class="payment-method-number text-muted">{{ method.maskedNumber }}</p>
            </div>
          </div>
          <div class="payment-method-actions">
            <v-btn class="action-btn">Edit</v-btn>
            <v-btn class="action-btn">Remove</v-btn>
          </div>
        </div>
      </div>
    </div>
    <v-btn class="action-btn add-payment-btn">
      <v-icon icon="mdi-plus" size="24" class="mr-2"></v-icon>
      Add Payment Method
    </v-btn>
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
.dashboard-card.payment-card {
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
.payment-methods {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.payment-method-item {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--rounded);
  padding: 1.2rem;
  transition: border-color var(--transition), background var(--transition);
  position: relative;
  overflow: hidden;
}
.payment-method-item:hover {
  border-color: var(--color-accent);
}
.payment-method-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
.payment-method-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  flex: 1;
}
.payment-method-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--rounded);
  background: none;
}
.payment-method-name {
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.1rem;
}
.payment-method-number {
  font-size: 0.9rem;
  color: var(--color-text-muted);
}
.payment-method-actions {
  display: flex;
  gap: 0.5rem;
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
.add-payment-btn {
  margin-top: 0.5rem;
  font-weight: 600;
  height: 40px;
}
</style>
