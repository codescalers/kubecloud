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
          { title: 'Balance', key: 'balance' },
          { title: 'Actions', key: 'actions', sortable: false, width: '160px' }
        ]"
        :items="users"
        :items-per-page="pageSize"
        :page="currentPage"
        @update:page="$emit('update:currentPage', $event)"
        class="admin-table"
        density="comfortable"
      >
        <template #item.balance="{ item }">
          ${{ item.balance.toFixed(2) }}
        </template>
        <template #item.actions="{ item }">
          <div style="display: flex; gap: var(--space-4); align-items: center;">
            <v-btn size="small" variant="outlined" class="action-btn" :disabled="!item.verified" @click="$emit('creditUser', item)">
              <v-icon icon="mdi-cash-plus" size="16" class="mr-1"></v-icon>
              Credit Balance
            </v-btn>
            <v-btn size="small" variant="outlined" class="action-btn" @click="$emit('deleteUser', item.id)">
              <v-icon icon="mdi-delete" size="16" class="mr-1"></v-icon>
              Remove
            </v-btn>
          </div>
        </template>
      </v-data-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { User } from '../stores/user'

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
