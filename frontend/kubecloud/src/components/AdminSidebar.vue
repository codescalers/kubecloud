<template>
  <v-card class="admin-sidebar-card cloud-card" flat>
    <v-list nav density="comfortable">
      <v-list-item
        v-for="item in adminNavItems"
        :key="item.key"
        :active="selected === item.key"
        @click="$emit('update:selected', item.key)"
        class="admin-sidebar-item"
        :class="{ 'admin-sidebar-item--active': selected === item.key }"
      >
        <template v-slot:prepend>
          <v-icon :icon="item.icon" size="24" color="accent" class="cloud-icon"></v-icon>
        </template>
        <v-list-item-title class="admin-sidebar-title">{{ item.title }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-card>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'
const props = defineProps<{ selected: string }>()
defineEmits(['update:selected'])

const adminNavItems = [
  { key: 'overview', title: 'Overview', icon: 'mdi-view-dashboard' },
  { key: 'users', title: 'Users', icon: 'mdi-account-group' },
  { key: 'vouchers', title: 'Vouchers', icon: 'mdi-ticket-percent' },
  { key: 'credits', title: 'Credits', icon: 'mdi-cash-plus' },
  { key: 'system', title: 'System', icon: 'mdi-cog' },
]
</script>

<style scoped>
.admin-sidebar-card {
  background: var(--glass-bg);
  border: 1.5px solid var(--color-accent);
  border-radius: 14px;
  box-shadow: var(--cloud-shadow);
  padding: 0.5rem 0.25rem;
}
.admin-sidebar-item {
  background: transparent;
  border-left: 3px solid transparent;
  border-radius: 8px;
  margin-bottom: 0.25rem;
  transition: background 0.2s, border-color 0.2s;
}
.admin-sidebar-item--active {
  background: var(--cloud-blue);
  color: var(--color-accent);
  border-left: 3px solid var(--color-accent);
}
.admin-sidebar-title {
  color: var(--color-text);
  font-weight: 600;
  font-size: 1rem;
}
.logout-item:hover {
  background: var(--color-error-bg, rgba(239,68,68,0.08));
}
</style> 