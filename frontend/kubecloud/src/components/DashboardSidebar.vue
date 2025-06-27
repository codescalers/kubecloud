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
  </v-card>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'

const props = defineProps<{ selected: string }>()
defineEmits(['update:selected'])

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
  background: rgba(10, 25, 47, 0.85);
  border: 1px solid rgba(96, 165, 250, 0.15);
  border-radius: 1rem;
  padding: .5rem;
  color: #CBD5E1;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-item {
  background: transparent;
  border: 1px solid transparent;
  margin-bottom: 0.25rem;
  min-height: 44px;
  padding: 0.75rem 1rem;
  border-radius: 0.75rem;
  transition: all 0.3s ease;
}

.sidebar-item:hover {
  background: rgba(96, 165, 250, 0.1);
  border: 1px solid rgba(96, 165, 250, 0.3);
}

.sidebar-item--active {
  background: rgba(96, 165, 250, 0.15) !important;
  border: 1px solid rgba(96, 165, 250, 0.5) !important;
  box-shadow: 0 0 15px rgba(96, 165, 250, 0.2);
}

.sidebar-icon {
  background: rgba(96, 165, 250, 0.1);
  border: 1px solid rgba(96, 165, 250, 0.2);
  border-radius: 0.5rem;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 0.75rem;
  transition: all 0.3s ease;
}

.sidebar-item:hover .sidebar-icon {
  background: rgba(96, 165, 250, 0.2);
  border: 1px solid rgba(96, 165, 250, 0.4);
  box-shadow: 0 0 8px rgba(96, 165, 250, 0.2);
}

.sidebar-item--active .sidebar-icon {
  background: rgba(96, 165, 250, 0.25) !important;
  border: 1px solid rgba(96, 165, 250, 0.6) !important;
  box-shadow: 0 0 12px rgba(96, 165, 250, 0.3) !important;
}

.sidebar-title {
  color: #CBD5E1;
  font-weight: 500;
  font-size: 0.95rem;
  transition: all 0.3s ease;
}

.sidebar-item:hover .sidebar-title {
  color: #60a5fa;
}

.sidebar-item--active .sidebar-title {
  color: #60a5fa !important;
  font-weight: 600;
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
