<template>
  <v-card class="pa-8 profile-card payment-card card-enhanced">
    <h2 class="display-3 mb-6 neon-glow-sky">Payment Methods</h2>
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
              <v-icon v-else size="32" :color="method.iconColor">{{ method.icon }}</v-icon>
            </div>
            <div class="payment-method-details">
              <h3 class="payment-method-name">{{ method.name }}</h3>
              <p class="payment-method-number">{{ method.maskedNumber }}</p>
            </div>
          </div>
          <div class="payment-method-actions">
            <v-btn variant="text" color="primary" size="large" class="action-btn neon-hover-violet">Edit</v-btn>
            <v-btn variant="text" color="error" size="large" class="action-btn neon-hover-rose">Remove</v-btn>
          </div>
        </div>
      </div>
    </div>
    <v-btn variant="elevated" color="primary" size="large" class="add-payment-btn neon-glow-sky">
      <v-icon icon="mdi-plus" size="24" class="mr-2"></v-icon>
      Add Payment Method
    </v-btn>
  </v-card>
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
.payment-card {
  border-radius: 24px;
  min-width: 100%;
}

.payment-methods {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.payment-method-item {
  background: rgba(124, 58, 237, 0.05);
  border: 1px solid rgba(124, 58, 237, 0.2);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.payment-method-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--kubecloud-purple), var(--kubecloud-purple-light));
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.payment-method-item:hover {
  background: rgba(124, 58, 237, 0.1);
  border: 1px solid var(--kubecloud-purple);
  box-shadow: var(--shadow-kubecloud-purple);
  transform: translateY(-1px);
}

.payment-method-item:hover::before {
  opacity: 1;
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
  gap: 1rem;
  flex: 1;
}

.payment-method-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: rgba(124, 58, 237, 0.1);
  border: 1px solid rgba(124, 58, 237, 0.3);
  transition: all var(--transition-normal);
}

.payment-method-item:hover .payment-method-icon {
  background: rgba(124, 58, 237, 0.2);
  border: 1px solid var(--kubecloud-purple);
  box-shadow: var(--shadow-kubecloud-purple);
  transform: scale(1.02);
}

.paypal-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: #0070BA;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.payment-method-item:hover .paypal-icon {
  color: #005EA6;
  transform: scale(1.05);
}

.payment-method-details {
  flex: 1;
}

.payment-method-name {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.25rem;
  font-weight: 600;
  color: #fff;
  margin: 0 0 0.25rem 0;
  letter-spacing: -0.01em;
}

.payment-method-number {
  font-size: 1rem;
  color: #E0E7EF;
  margin: 0;
  font-weight: 400;
  letter-spacing: 0.02em;
}

.payment-method-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.action-btn {
  font-weight: 600;
  font-size: 1rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: transparent;
  border: 1px solid transparent;
  transition: all var(--transition-normal);
  color: #fff !important;
}

.action-btn:hover {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

.action-btn.neon-hover-rose:hover {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--kubecloud-error);
  box-shadow: var(--shadow-kubecloud-error);
  text-shadow: 0 0 1px rgba(239, 68, 68, 0.3);
  transform: translateY(-1px);
}

.add-payment-btn {
  font-weight: 600;
  font-size: 1.125rem;
  padding: 1rem 1.5rem;
  border-radius: 12px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: linear-gradient(135deg, var(--kubecloud-blue), var(--kubecloud-blue-light)) !important;
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  color: var(--kubecloud-white) !important;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
}

.add-payment-btn:hover {
  background: linear-gradient(135deg, var(--kubecloud-blue-light), var(--kubecloud-blue)) !important;
  box-shadow: var(--shadow-hover-blue);
  transform: translateY(-2px);
}

@media (max-width: 960px) {
  .payment-card {
    padding: 2rem !important;
  }
  
  .payment-method-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .payment-method-actions {
    align-self: flex-end;
  }
  
  .add-payment-btn {
    font-size: 1rem;
    padding: 0.75rem 1.25rem;
    min-height: 44px;
  }
}

@media (max-width: 600px) {
  .payment-card {
    padding: 1.5rem !important;
  }
  
  .payment-method-item {
    padding: 1rem;
  }
  
  .payment-method-icon {
    width: 48px;
    height: 48px;
  }
  
  .payment-method-name {
    font-size: 1.125rem;
  }
  
  .payment-method-number {
    font-size: 0.9rem;
  }
  
  .action-btn {
    font-size: 0.9rem;
    padding: 0.375rem 0.75rem;
  }
  
  .add-payment-btn {
    font-size: 0.9rem;
    padding: 0.5rem 1rem;
    min-height: 40px;
  }
}
</style>
