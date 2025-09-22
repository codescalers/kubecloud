<template>
  <div class="notification-bell">
    <v-menu
      v-model="showDropdown"
      :close-on-content-click="false"
      location="bottom end"
      :offset="[0, 12]"
      width="400"
      :z-index="9999"
      transition="slide-y-transition"
    >
      <template v-slot:activator="{ props }">
        <v-btn
          icon
          variant="text"
          color="white"
          class="mx-1"
          v-bind="props"
        >
          <v-badge
            :content="unreadCount > 99 ? '99+' : unreadCount.toString()"
            :model-value="unreadCount > 0"
            color="error"
            offset-x="8"
            offset-y="-8"
          >
            <v-icon icon="mdi-bell" size="24"></v-icon>
          </v-badge>
        </v-btn>
      </template>

      <v-card class="notification-dropdown">
        <v-card-title class="d-flex align-center justify-space-between pa-4 bg-primary text-white">
          <span class="text-h6 font-weight-medium">Notifications</span>
          <v-btn
            v-if="unreadCount > 0"
            size="small"
            variant="text"
            color="white"
            @click="markAllAsRead"
            :loading="loading"
            class="text-caption"
          >
            Mark all read
          </v-btn>
        </v-card-title>

        <v-divider></v-divider>

        <div class="notification-list">
          <div v-if="loading && persistentNotifications.length === 0" class="pa-4 text-center">
            <v-progress-circular indeterminate color="primary"></v-progress-circular>
            <div class="mt-2 text-body-2 text-medium-emphasis">Loading notifications...</div>
          </div>

          <div v-else-if="persistentNotifications.length === 0" class="pa-4 text-center">
            <v-icon icon="mdi-bell-off" size="48" color="grey-lighten-1"></v-icon>
            <div class="mt-2 text-body-2 text-medium-emphasis">No notifications yet</div>
          </div>

          <div v-else>
            <v-list>
              <v-list-item
                v-for="notification in displayedNotifications"
                :key="notification.id"
                :class="[
                  'cursor-pointer',
                  'py-2',
                ]"
                :style="notification.status === 'unread' ? unreadItemStyle : readItemStyle"
                @click="onNotificationClick(notification)"
                :ripple="true"
              >
                <template v-slot:prepend>
                  <v-avatar size="40" :color="getNotificationColor(notification.severity)" class="notification-icon">
                    <v-icon :icon="getNotificationIcon(notification.type)" color="white"></v-icon>
                  </v-avatar>
                </template>

                <v-list-item-title class="text-body-2 font-weight-medium">{{ notification.payload.subject ||
                  capitalize(notification.type) || 'Notification'
                  }}</v-list-item-title>
                <v-list-item-subtitle class="text-caption">{{ notification.payload.message || notification.payload.description || notification.payload.details || '' }}</v-list-item-subtitle>
                <v-list-item-subtitle class="text-caption text-medium-emphasis">{{ formatNotificationTime(notification.created_at) }}</v-list-item-subtitle>

                <template v-slot:append>
                  <v-btn
                    v-if="notification.status === 'unread'"
                    icon
                    size="small"
                    variant="text"
                    @click.stop="markAsRead(notification.id)"
                    color="success"
                  >
                    <v-icon icon="mdi-check" size="16"></v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
          </div>

          <v-divider v-if="persistentNotifications.length > 0"></v-divider>

          <div v-if="persistentNotifications.length > 0" class="pa-3 text-center">
            <v-btn
              variant="text"
              color="primary"
              size="small"
              @click="navigateToNotificationsPage"
              prepend-icon="mdi-eye"
            >
              See All Notifications
            </v-btn>
          </div>
        </div>
      </v-card>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, capitalize } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useNotificationStore } from '../stores/notifications'
import type { Notification } from '../types/notifications'
import { getNotificationIcon, getNotificationColor, formatNotificationTime } from '../utils/notificationUtils'

const router = useRouter()
const notificationStore = useNotificationStore()
const showDropdown = ref(false)
const dropdownLimit = 10

// Use storeToRefs to maintain reactivity
const {
  persistentNotifications,
  loading,
  unreadCount
} = storeToRefs(notificationStore)

// Destructure methods (these don't need reactivity)
const {
  markAsRead,
  markAllAsRead,
  loadNotifications
} = notificationStore

// Computed property to limit displayed notifications
const displayedNotifications = computed(() =>
  persistentNotifications.value.slice(0, dropdownLimit)
)

const onNotificationClick = (notification: Notification) => {
  if (notification.status === 'unread') {
    markAsRead(notification.id)
  }
  showDropdown.value = false
  router.push('/notifications')
}


const navigateToNotificationsPage = () => {
  showDropdown.value = false
  router.push('/notifications')
}

// Watch for dropdown state changes
watch(showDropdown, (newValue) => {
  if (newValue && persistentNotifications.value.length === 0) {
    loadNotifications()
  }
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