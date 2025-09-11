<template>
  <div>
    <div class="container mx-auto pa-6 mt-12">
      <!-- Header -->
      <div class="d-flex align-center justify-space-between mb-6">
        <div>
          <h1 class="text-h4 font-weight-bold mb-2">Notifications</h1>
          <p class="text-body-1 text-medium-emphasis">
            Manage and view all your notifications
          </p>
        </div>

        <div class="d-flex gap-3">
          <v-btn
            v-if="unreadCount > 0"
            variant="outlined"
            color="primary"
            @click="markAllAsRead"
            :loading="loading"
            prepend-icon="mdi-check-all"
          >
            Mark All Read
          </v-btn>

          <v-btn
            variant="outlined"
            color="secondary"
            @click="openClearAllDialog"
            :loading="loading"
            :disabled="persistentNotifications.length === 0 || loading"
            prepend-icon="mdi-delete-sweep"
          >
            Clear All
          </v-btn>
        </div>
      </div>

      <!-- Filters -->
      <div class="d-flex align-center gap-4 mb-5">
        <v-btn-toggle
          v-model="statusFilter"
          mandatory
          color="primary"
          variant="outlined"
        >
          <v-btn value="all">All</v-btn>
          <v-btn value="unread">Unread ({{ unreadCount }})</v-btn>
          <v-btn value="read">Read</v-btn>
        </v-btn-toggle>

        <v-select
          v-model="typeFilter"
          :items="typeOptions"
          label="Type"
          variant="outlined"
          density="compact"
          hide-details
          class="ml-4"
          style="min-width: 200px"
        ></v-select>
      </div>


      <!-- Notifications List -->
      <v-card elevation="2">
        <v-card-text class="pa-0">
          <div v-if="loading && persistentNotifications.length === 0" class="pa-8 text-center">
            <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
            <div class="mt-4 text-h6 text-medium-emphasis">Loading notifications...</div>
          </div>

          <div v-else-if="filteredNotifications.length === 0" class="pa-8 text-center">
            <v-icon icon="mdi-bell-off" size="80" color="grey-lighten-1"></v-icon>
            <div class="mt-4 text-h6 text-medium-emphasis">No notifications found</div>
            <p class="text-body-1 text-grey mt-2">
              {{ statusFilter === 'all' ? 'You have no notifications yet.' : `No ${statusFilter} notifications found.` }}
            </p>
          </div>

          <div v-else>
            <v-list>
              <v-list-item
                v-for="notification in paginatedNotifications"
                :key="notification.id"
                :class="[
                  'cursor-pointer',
                  'pa-4',
                  'mb-3',
                  'mx-2',
                  'rounded-lg',
                  'elevation-2',
                ]"
                :style="notification.status === 'unread' ? unreadItemStyle : readItemStyle"
                @click="onNotificationClick(notification)"
                :ripple="true"
              >
                <template v-slot:prepend>
                  <v-avatar
                    size="48"
                    :color="getNotificationColor(notification.severity)"
                    class="notification-icon mr-4"
                  >
                    <v-icon
                      :icon="getNotificationIcon(notification.severity)"
                      color="white"
                      size="24"
                    ></v-icon>
                  </v-avatar>
                </template>

                <v-list-item-title class="text-h6 font-weight-medium mb-2">
                  {{ notification.payload.title || notification.payload.message || 'Notification' }}
                </v-list-item-title>

                <v-list-item-subtitle class="text-body-1 mb-2">
                  {{ notification.payload.message || notification.payload.description || notification.payload.details || '' }}
                </v-list-item-subtitle>

                <div class="d-flex align-center justify-space-between">
                  <v-list-item-subtitle class="text-caption text-medium-emphasis">
                    {{ formatNotificationTime(notification.created_at) }}
                  </v-list-item-subtitle>

                  <div class="d-flex gap-2 align-center">
                    <v-chip
                      :color="getNotificationColor(notification.severity)"
                      variant="tonal"
                      size="small"
                      class="text-caption"
                    >
                      {{ notification.severity.toUpperCase() }}
                    </v-chip>

                    <v-chip
                      v-if="notification.status === 'unread'"
                      color="primary"
                      variant="tonal"
                      size="small"
                      class="text-caption"
                    >
                      UNREAD
                    </v-chip>

                    <v-btn
                      v-if="notification.status === 'unread'"
                      icon
                      size="small"
                      variant="text"
                      @click.stop="markAsRead(notification.id)"
                      color="success"
                      class="ml-2"
                    >
                      <v-icon icon="mdi-check" size="16"></v-icon>
                    </v-btn>

                    <v-btn
                      v-else
                      icon
                      size="small"
                      variant="text"
                      @click.stop="markAsUnread(notification.id)"
                      color="primary"
                      class="ml-2"
                    >
                      <v-icon icon="mdi-email-outline" size="16"></v-icon>
                    </v-btn>

                    <v-btn
                      icon
                      size="small"
                      variant="text"
                      @click.stop="openDeleteDialog(notification)"
                      color="error"
                      class="ml-1"
                    >
                      <v-icon icon="mdi-delete" size="16"></v-icon>
                    </v-btn>
                  </div>
                </div>
              </v-list-item>
            </v-list>

            <!-- Pagination -->
            <v-divider></v-divider>
            <div class="d-flex align-center justify-space-between pa-4">
              <div class="text-body-2 text-medium-emphasis">
                Showing {{ startIndex + 1 }}-{{ endIndex }} of {{ filteredNotifications.length }} notifications
              </div>

              <v-pagination
                v-model="currentPage"
                :length="totalPages"
                :total-visible="7"
                color="primary"
              ></v-pagination>
            </div>
          </div>
        </v-card-text>
      </v-card>
      
      <!-- Delete Confirmation Dialog -->
      <v-dialog v-model="showDeleteDialog" max-width="420">
        <v-card class="pa-3">
          <v-card-title class="text-h6">Delete Notification</v-card-title>
          <v-card-text>
            Are you sure you want to delete this notification? This action cannot be undone.
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn variant="text" color="grey" @click="showDeleteDialog = false">Cancel</v-btn>
            <v-btn variant="outlined" color="error" :loading="deleting" @click="confirmDelete">Delete</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Clear All Confirmation Dialog -->
      <v-dialog v-model="showClearAllDialog" max-width="420">
        <v-card class="pa-3">
          <v-card-title class="text-h6">Clear All Notifications</v-card-title>
          <v-card-text>
            Are you sure you want to clear all notifications? This action cannot be undone.
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn variant="text" color="grey" @click="showClearAllDialog = false">Cancel</v-btn>
            <v-btn variant="outlined" color="error" :loading="clearing" @click="confirmClearAll">Clear All</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useNotificationStore, type Notification } from '../stores/notifications'
import { getNotificationIcon, getNotificationColor, formatNotificationTime } from '../utils/notificationUtils'

const notificationStore = useNotificationStore()

// Use storeToRefs to maintain reactivity
const {
  persistentNotifications,
  loading,
  unreadCount
} = storeToRefs(notificationStore)

// Destructure methods (these don't need reactivity)
const {
  markAsRead,
  markAsUnread,
  markAllAsRead,
  clearAll,
  deleteNotification,
  loadNotifications
} = notificationStore

// Filters
const statusFilter = ref<'all' | 'read' | 'unread'>('all')
const typeFilter = ref<'all' | 'deployment' | 'billing' | 'user' | 'connected'>('all')
const currentPage = ref(1)
const pageSize = 10

// Type options for filter
const typeOptions = computed(() => [
  { title: 'All Types', value: 'all' },
  { title: 'Deployment', value: 'deployment' },
  { title: 'Billing', value: 'billing' },
  { title: 'User', value: 'user' },
  { title: 'Connected', value: 'connected' }
])

// Filtered notifications
const filteredNotifications = computed(() => {
  let filtered = [...persistentNotifications.value]

  // Filter by status
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter((n: Notification) => n.status === statusFilter.value as 'read' | 'unread')
  }

  // Filter by type
  if (typeFilter.value !== 'all') {
    filtered = filtered.filter((n: Notification) => n.type === typeFilter.value)
  }

  return filtered
})

// Pagination
const totalPages = computed(() => Math.ceil(filteredNotifications.value.length / pageSize))
const startIndex = computed(() => (currentPage.value - 1) * pageSize)
const endIndex = computed(() => Math.min(startIndex.value + pageSize, filteredNotifications.value.length))
const paginatedNotifications = computed(() =>
  filteredNotifications.value.slice(startIndex.value, endIndex.value)
)

// Methods
const onNotificationClick = async (notification: Notification) => {
  if (notification.status === 'unread') {
    await notificationStore.markAsRead(notification.id)
  }
}

// Dialog state and handlers
const showClearAllDialog = ref(false)
const showDeleteDialog = ref(false)
const notificationToDelete = ref<Notification | null>(null)
const clearing = ref(false)
const deleting = ref(false)

const openClearAllDialog = () => {
  showClearAllDialog.value = true
}

const openDeleteDialog = (notification: Notification) => {
  notificationToDelete.value = notification
  showDeleteDialog.value = true
}

const confirmClearAll = async () => {
  try {
    clearing.value = true
    await clearAll()
  } finally {
    clearing.value = false
    showClearAllDialog.value = false
  }
}

const confirmDelete = async () => {
  if (!notificationToDelete.value) return
  try {
    deleting.value = true
    await deleteNotification(notificationToDelete.value.id)
  } finally {
    deleting.value = false
    showDeleteDialog.value = false
    notificationToDelete.value = null
  }
}

// Watch for filter changes
watch([statusFilter, typeFilter], () => {
  currentPage.value = 1
})

// Initial load
onMounted(() => {
  if (persistentNotifications.value.length === 0) {
    loadNotifications()
  }
})

// Minimal inline styles for gradient backgrounds
const unreadItemStyle = {
  background: 'linear-gradient(135deg, var(--color-bg-elevated) 0%, var(--color-bg-hover) 100%)',
  borderLeft: '4px solid var(--color-primary)'
} as const

const readItemStyle = {
  background: 'linear-gradient(135deg, var(--color-bg) 0%, var(--color-bg-elevated) 100%)',
  borderLeft: '4px solid var(--color-border)'
} as const
</script>

<style scoped>
</style>