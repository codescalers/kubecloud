<template>
  <v-card class="pa-8 profile-card vouchers-card card-enhanced">
    <h2 class="display-3 mb-6 kubecloud-gradient">Vouchers & Credits</h2>
    <div class="vouchers-list mb-6">
      <div v-for="voucher in vouchers" :key="voucher.id" class="voucher-item">
        <div class="voucher-content">
          <div class="voucher-info">
            <div class="voucher-icon">
              <v-icon size="32" :color="voucher.iconColor">{{ voucher.icon }}</v-icon>
            </div>
            <div class="voucher-details">
              <h3 class="voucher-name">{{ voucher.name }}</h3>
              <p class="voucher-description">{{ voucher.description }}</p>
              <p class="voucher-expiry">Expires {{ voucher.expiryDate }}</p>
            </div>
          </div>
          <div class="voucher-value">
            <span class="voucher-amount">{{ voucher.amount }}</span>
            <v-btn variant="text" color="primary" size="large" class="redeem-btn kubecloud-hover-blue">Redeem</v-btn>
          </div>
        </div>
      </div>
    </div>
    <v-btn variant="elevated" color="primary" size="large" class="add-voucher-btn btn-enhanced">
      <v-icon icon="mdi-plus" size="24" class="mr-2"></v-icon>
      Add Voucher Code
    </v-btn>
  </v-card>
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
.vouchers-card {
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 1px solid var(--glass-border);
  border-radius: 24px;
  box-shadow: var(--shadow-md), var(--shadow-ambient);
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
  min-width: 100%;
}

.vouchers-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--kubecloud-orange-subtle), transparent);
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.vouchers-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl), var(--shadow-ambient);
  border-color: var(--kubecloud-orange-subtle);
}

.vouchers-card:hover::before {
  opacity: 1;
}

.vouchers-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.voucher-item {
  background: rgba(234, 88, 12, 0.05);
  border: 1px solid rgba(234, 88, 12, 0.2);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.voucher-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--kubecloud-orange), var(--kubecloud-orange-light));
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.voucher-item:hover {
  background: rgba(234, 88, 12, 0.1);
  border: 1px solid var(--kubecloud-orange);
  box-shadow: var(--shadow-kubecloud-orange);
  transform: translateY(-1px);
}

.voucher-item:hover::before {
  opacity: 1;
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
  gap: 1rem;
  flex: 1;
}

.voucher-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: rgba(234, 88, 12, 0.1);
  border: 1px solid rgba(234, 88, 12, 0.3);
  transition: all var(--transition-normal);
}

.voucher-item:hover .voucher-icon {
  background: rgba(234, 88, 12, 0.2);
  border: 1px solid var(--kubecloud-orange);
  box-shadow: var(--shadow-kubecloud-orange);
  transform: scale(1.02);
}

.voucher-details {
  flex: 1;
}

.voucher-name {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.25rem;
  font-weight: 600;
  color: #fff;
  margin: 0 0 0.25rem 0;
  letter-spacing: -0.01em;
}

.voucher-description {
  font-size: 1rem;
  color: #E0E7EF;
  margin: 0 0 0.25rem 0;
  font-weight: 400;
  opacity: 0.9;
}

.voucher-expiry {
  font-size: 0.875rem;
  color: var(--kubecloud-orange);
  margin: 0;
  font-weight: 500;
  opacity: 0.8;
  text-shadow: 0 0 1px rgba(234, 88, 12, 0.3);
}

.voucher-value {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.75rem;
}

.voucher-amount {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--kubecloud-orange);
  text-shadow: 0 0 1px rgba(234, 88, 12, 0.3);
  letter-spacing: -0.02em;
}

.redeem-btn {
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

.redeem-btn:hover {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

.add-voucher-btn {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-weight: 600;
  font-size: 1.125rem;
  padding: 1rem 2rem;
  border-radius: 12px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: linear-gradient(135deg, var(--kubecloud-blue), var(--kubecloud-blue-dark)) !important;
  box-shadow: var(--shadow-kubecloud-blue);
  transition: all var(--transition-normal);
  color: #fff !important;
  width: 100%;
  max-width: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 56px;
}

.add-voucher-btn:hover {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

@media (max-width: 960px) {
  .vouchers-card {
    padding: 2rem !important;
  }
  .voucher-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  .voucher-value {
    width: 100%;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
  .voucher-name {
    font-size: 1.125rem;
  }
  .voucher-description {
    font-size: 0.9rem;
  }
  .voucher-expiry {
    font-size: 0.8rem;
  }
  .voucher-amount {
    font-size: 1.25rem;
  }
  .redeem-btn {
    padding: 0.375rem 0.75rem;
    font-size: 0.9rem;
  }
  .add-voucher-btn {
    padding: 0.875rem 1.5rem;
    font-size: 1rem;
  }
}

@media (max-width: 600px) {
  .vouchers-card {
    padding: 1.5rem !important;
  }
  .voucher-item {
    padding: 1rem;
  }
  .voucher-icon {
    width: 48px;
    height: 48px;
  }
  .voucher-name {
    font-size: 1rem;
  }
  .voucher-description {
    font-size: 0.875rem;
  }
  .voucher-expiry {
    font-size: 0.75rem;
  }
  .voucher-amount {
    font-size: 1.125rem;
  }
  .redeem-btn {
    padding: 0.25rem 0.5rem;
    font-size: 0.8rem;
  }
  .add-voucher-btn {
    padding: 0.75rem 1rem;
    font-size: 0.9rem;
  }
}

.text-success {
  color: var(--kubecloud-success) !important;
  text-shadow: 0 0 1px rgba(16, 185, 129, 0.3);
}
</style>
