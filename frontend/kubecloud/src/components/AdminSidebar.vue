<template>
  <div class="admin-sidebar-card">
    <div class="sidebar-header">
      <h3 class="sidebar-title">Admin Panel</h3>
      <p class="sidebar-subtitle">Platform management tools</p>
    </div>
    <v-list nav density="comfortable" class="sidebar-list">
      <v-list-item
        v-for="item in adminNavItems"
        :key="item.key"
        :active="selected === item.key"
        @click="$emit('update:selected', item.key)"
        class="sidebar-item"
        :class="{ 'sidebar-item--active': selected === item.key }"
      >
        <template v-slot:prepend>
          <div class="sidebar-icon">
            <v-icon :icon="item.icon" size="20" :color="selected === item.key ? 'var(--color-primary)' : 'var(--color-text-secondary)'"></v-icon>
          </div>
        </template>
        <v-list-item-title class="sidebar-item-title">{{ item.title }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, onMounted } from 'vue'
const props = defineProps<{ selected: string }>()
defineEmits(['update:selected'])

const adminNavItems = [
  { key: 'overview', title: 'Overview', icon: 'mdi-view-dashboard' },
  { key: 'users', title: 'Users', icon: 'mdi-account-group' },
  { key: 'vouchers', title: 'Vouchers', icon: 'mdi-ticket-percent' },
  { key: 'credits', title: 'Credits', icon: 'mdi-cash-plus' },
  { key: 'system', title: 'System', icon: 'mdi-cog' },
]

onMounted(() => {
  console.log('AdminSidebar mounted, selected:', props.selected)
})
</script>

<style scoped>
.admin-sidebar-card {
  transition: all var(--transition-normal, 0.2s);
  backdrop-filter: blur(8px);
}

.sidebar-header {
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border, #334155);
}

.sidebar-title {
  font-size: var(--font-size-lg, 1.125rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text, #F8FAFC);
  margin: 0 0 0.25rem 0;
}

.sidebar-subtitle {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-text-secondary, #CBD5E1);
  margin: 0;
}

.sidebar-list {
  background: transparent !important;
}

.sidebar-item {
  background: transparent !important;
  border-radius: var(--radius-md, 0.375rem);
  margin-bottom: 0.25rem;
  transition: all var(--transition-normal, 0.2s);
  border: 1px solid transparent;
}

.sidebar-item:hover {
  background: rgba(30, 41, 59, 0.3) !important;
  border-color: var(--color-border-light, #475569);
}

.sidebar-item--active {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: var(--color-primary, #3B82F6);
  border-left: 3px solid var(--color-primary, #3B82F6);
}

.sidebar-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md, 0.375rem);
  background: rgba(59, 130, 246, 0.05);
  border: 1px solid transparent;
  transition: all var(--transition-normal, 0.2s);
}

.sidebar-item--active .sidebar-icon {
  background: rgba(59, 130, 246, 0.15);
  border-color: var(--color-primary, #3B82F6);
}

.sidebar-item-title {
  color: var(--color-text, #F8FAFC);
  font-weight: var(--font-weight-medium, 500);
  font-size: var(--font-size-base, 1rem);
}

.sidebar-item--active .sidebar-item-title {
  color: var(--color-primary, #3B82F6);
  font-weight: var(--font-weight-semibold, 600);
}

/* Vuetify overrides */
.sidebar-list :deep(.v-list-item) {
  background: transparent !important;
  box-shadow: none !important;
  border: none !important;
  color: inherit !important;
  min-height: 44px;
  padding: 0.5rem 0.75rem;
}

.sidebar-list :deep(.v-list-item__prepend) {
  margin-right: 0.75rem !important;
}

.sidebar-list :deep(.v-list-item--active) {
  background: transparent !important;
}

/* Responsive */
@media (max-width: 900px) {
  .admin-sidebar-card {
    padding: 1rem;
  }
  
  .sidebar-header {
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
  }
  
  .sidebar-title {
    font-size: var(--font-size-base, 1rem);
  }
  
  .sidebar-subtitle {
    font-size: var(--font-size-xs, 0.75rem);
  }
}
</style> 