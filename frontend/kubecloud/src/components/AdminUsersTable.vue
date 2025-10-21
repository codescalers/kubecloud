<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">User Search</h3>
      <p class="dashboard-card-subtitle">Find and manage existing users</p>
    </div>
    <v-text-field
      v-model="searchQueryLocal"
      label="Search users by name or email"
      prepend-inner-icon="mdi-magnify"
      variant="outlined"
      density="comfortable"
      clearable
      class="search-field"
      @input="$emit('update:searchQuery', searchQueryLocal)"
    />
    <div class="table-container">
      <v-data-table
        :headers="[
          { title: 'ID', key: 'id', width: '80px' },
          { title: 'Name', key: 'username' },
          { title: 'Email', key: 'email' },
          { title: 'Balance (USD)', key: 'balance_in_usd' },
          { title: 'Balance (TFT)', key: 'balance_in_tft' },
          { title: 'Admin', key: 'admin' },
          { title: 'Actions', key: 'actions', sortable: false, width: '160px' }
        ]"
        :items="users"
        :items-per-page="pageSize"
        :page="currentPage"
        :sort-by="[{ key: 'id', order: 'asc' }]"
        @update:page="$emit('update:currentPage', $event)"
        class="admin-table"
        density="comfortable"
      >
        <template #item.balance_in_usd="{ item }">
          ${{ +calculateNetBalance(item).toFixed(2) }}
        </template>
        <template #item.balance_in_tft="{ item }">
          {{ +item.balance_in_tft.toFixed(2) }}
        </template>
        <template #item.admin="{ item }">
          <v-checkbox v-if="item.admin" style=" display: flex; align-items: center;" :model-value="item.admin" disabled></v-checkbox>
        </template>
        <template #item.actions="{ item }">
          <div style="display: flex; gap: var(--space-4); align-items: center;">
            <v-tooltip location="bottom" text="Credit user">
              <template #activator="{ props }">
                <v-btn v-bind="props" size="small" icon="mdi-cash-plus" :disabled="!item.verified" @click="$emit('creditUser', item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip location="bottom" text="Delete user">
              <template #activator="{ props }">
                <v-btn v-bind="props" size="small" icon="mdi-delete" @click="$emit('deleteUser', item.id)"></v-btn>
              </template>
            </v-tooltip>
          </div>
        </template>
      </v-data-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { User } from '../types/user'

import { calculateNetBalance } from '../utils/dateUtils'
const props = defineProps({
  users: Array as () => User[],
  searchQuery: String,
  currentPage: Number,
  pageSize: Number
})
const emit = defineEmits(['update:searchQuery', 'update:currentPage', 'deleteUser', 'creditUser'])

const searchQueryLocal = ref(props.searchQuery)

watch(() => props.searchQuery, (val) => { searchQueryLocal.value = val })
</script>
