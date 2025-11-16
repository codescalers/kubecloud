<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">User Search</h3>
      <p class="dashboard-card-subtitle">Find and manage existing users</p>
    </div>
    <div style="display: flex; gap: 1rem; margin-bottom: 1.5rem; align-items: center; width: 100%;">
      <div style="flex: 1; min-width: 0;">
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
      </div>
      <v-btn
        color="warning"
        variant="outlined"
        prepend-icon="mdi-water-remove"
        class="drain-all-btn"
        @click="$emit('drainAllUsers')"
      >
        Drain All Users
      </v-btn>
    </div>
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
          ${{ item.balance != null ? item.balance.toFixed(2) : 'N/A' }}
        </template>
        <template #item.actions="{ item }">
          <div style="display: flex; gap: var(--space-4); align-items: center;">
            <v-btn size="small" variant="outlined" class="action-btn" :disabled="!item.verified" @click="$emit('creditUser', item)">
              <v-icon icon="mdi-cash-plus" size="16" class="mr-1"></v-icon>
              Credit Balance
            </v-btn>
            <v-btn size="small" variant="outlined" class="action-btn" color="warning" @click="$emit('drainUser', item)">
              <v-icon icon="mdi-water-remove" size="16" class="mr-1"></v-icon>
              Drain
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
const emit = defineEmits(['update:searchQuery', 'update:currentPage', 'deleteUser', 'creditUser', 'drainUser', 'drainAllUsers'])

const searchQueryLocal = ref(props.searchQuery)

watch(() => props.searchQuery, (val) => { searchQueryLocal.value = val })
</script>

<style scoped>
.drain-all-btn {
  white-space: nowrap;
  flex-shrink: 0;
}
</style>
