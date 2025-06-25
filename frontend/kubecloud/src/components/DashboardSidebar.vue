<template>
  <v-card class="sidebar-card" flat>
    <v-list nav density="comfortable">
      <v-list-item 
        v-for="item in navigationItems"
        :key="item.key"
        :active="selected === item.key" 
        @click="$emit('update:selected', item.key)" 
        class="sidebar-item"
        :class="{ 'sidebar-item--active': selected === item.key }"
      >
        <template v-slot:prepend>
          <div class="sidebar-icon">
            <v-icon :icon="item.icon" size="24" color="primary"></v-icon>
          </div>
        </template>
        <v-list-item-title class="sidebar-title">{{ item.title }}</v-list-item-title>
      </v-list-item>
    </v-list>
    
    <v-divider class="my-4 sidebar-divider" />
    
    <v-list nav density="comfortable">
      <v-list-item @click="$emit('logout')" class="sidebar-item logout-item">
        <template v-slot:prepend>
          <div class="sidebar-icon logout-icon">
            <v-icon icon="mdi-logout" size="24" color="error"></v-icon>
          </div>
        </template>
        <v-list-item-title class="sidebar-title text-error">Logout</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-card>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'

const props = defineProps<{ selected: string }>()
defineEmits(['update:selected', 'logout'])

// Navigation items data
const navigationItems = [
  {
    key: 'overview',
    title: 'Overview',
    icon: 'mdi-view-dashboard'
  },
  {
    key: 'clusters',
    title: 'Clusters',
    icon: 'mdi-server'
  },
  {
    key: 'profile',
    title: 'Profile',
    icon: 'mdi-account'
  },
  {
    key: 'ssh',
    title: 'SSH Keys',
    icon: 'mdi-key'
  },
  {
    key: 'payment',
    title: 'Payment',
    icon: 'mdi-credit-card'
  },
  {
    key: 'billing',
    title: 'Billing History',
    icon: 'mdi-receipt'
  },
  {
    key: 'vouchers',
    title: 'Vouchers',
    icon: 'mdi-ticket-percent'
  }
]
</script>

<style scoped>
.sidebar-card {
  border-radius: 18px;
  box-shadow: 0 2px 12px rgba(60,60,60,0.06);
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 1px solid var(--kubecloud-blue);
  padding: 1.5rem 1rem;
  height: 100%;
}

.sidebar-item {
  border-radius: 12px;
  margin-bottom: 0.5rem;
  min-height: 48px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: transparent;
  border: 1px solid transparent;
}

.sidebar-item:hover {
  background: rgba(37, 99, 235, 0.1);
  border: 1px solid var(--kubecloud-blue);
  transform: translateX(2px);
}

.sidebar-item--active {
  background: rgba(37, 99, 235, 0.15) !important;
  border: 1px solid var(--kubecloud-blue) !important;
  box-shadow: 0 0 15px rgba(37, 99, 235, 0.2);
}

.sidebar-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(37, 99, 235, 0.1);
  border: 1px solid rgba(37, 99, 235, 0.2);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-right: 12px;
}

.sidebar-item:hover .sidebar-icon {
  background: rgba(37, 99, 235, 0.2);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: 0 0 8px rgba(37, 99, 235, 0.2);
  transform: scale(1.02);
}

.sidebar-item--active .sidebar-icon {
  background: rgba(37, 99, 235, 0.25) !important;
  border: 1px solid var(--kubecloud-blue) !important;
  box-shadow: 0 0 12px rgba(37, 99, 235, 0.3) !important;
}

.sidebar-title {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1rem;
  font-weight: 500;
  color: #fff;
  letter-spacing: 0.01em;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-item:hover .sidebar-title {
  color: var(--kubecloud-blue);
  text-shadow: 0 0 1px rgba(37, 99, 235, 0.3);
}

.sidebar-item--active .sidebar-title {
  color: var(--kubecloud-blue) !important;
  font-weight: 600 !important;
  text-shadow: 0 0 1px rgba(37, 99, 235, 0.3) !important;
}

.sidebar-divider {
  border-color: rgba(37, 99, 235, 0.2);
  margin: 1.5rem 0;
}

.logout-item:hover {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--neon-rose);
}

.logout-item:hover .sidebar-icon {
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid var(--neon-rose);
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.2);
}

.logout-item:hover .sidebar-title {
  color: var(--neon-rose);
  text-shadow: 0 0 1px rgba(239, 68, 68, 0.3);
}

.logout-icon {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

@media (max-width: 960px) {
  .sidebar-card {
    padding: 1rem 0.75rem;
  }
  
  .sidebar-icon {
    width: 36px;
    height: 36px;
    margin-right: 10px;
  }
  
  .sidebar-title {
    font-size: 0.9rem;
  }
}

@media (max-width: 600px) {
  .sidebar-card {
    padding: 0.75rem 0.5rem;
  }
  
  .sidebar-item {
    margin-bottom: 0.25rem;
    min-height: 44px;
  }
  
  .sidebar-icon {
    width: 32px;
    height: 32px;
    margin-right: 8px;
  }
  
  .sidebar-title {
    font-size: 0.85rem;
  }
}
</style>
