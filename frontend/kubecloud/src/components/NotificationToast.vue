<template>
  <div class="notification-container">
    <transition-group name="toast-fade-slide" tag="div">
      <div 
        v-for="notification in notifications" 
        :key="notification.id" 
        :class="['toast', notification.type]"
      >
        <v-icon class="toast-icon" :color="getIconColor(notification.type)" left>
          {{ getIcon(notification.type) }}
        </v-icon>
        <div class="toast-content">
          <div class="toast-title">{{ notification.title }}</div>
          <div class="toast-message">{{ notification.message }}</div>
        </div>
        <v-btn icon class="toast-close" @click="removeNotification(notification.id)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useNotificationStore } from '../stores/notifications'
import { storeToRefs } from 'pinia'

const notificationStore = useNotificationStore()
const { notifications } = storeToRefs(notificationStore)

const getIcon = (type: string) => {
  switch (type) {
    case 'success': return 'mdi-check-circle'
    case 'error': return 'mdi-alert-circle'
    case 'warning': return 'mdi-alert'
    case 'info': return 'mdi-information'
    default: return 'mdi-information'
  }
}

const getIconColor = (type: string) => {
  switch (type) {
    case 'success': return '#10B981'
    case 'error': return '#EF4444'
    case 'warning': return '#F59E0B'
    case 'info': return '#60a5fa'
    default: return '#60a5fa'
  }
}

const removeNotification = (id: string) => {
  notificationStore.removeNotification(id)
}
</script>

<style scoped>
.notification-container {
  position: fixed;
  top: 2rem;
  right: 2rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: flex-start;
  background: rgba(32, 38, 50, 0.95);
  color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  padding: 1rem 1.5rem;
  min-width: 320px;
  max-width: 450px;
  font-weight: 500;
  font-size: 0.95rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  pointer-events: auto;
  animation: toast-fade-in 0.3s ease-out;
}

.toast-content {
  flex: 1;
  margin: 0 1rem;
}

.toast-title {
  font-weight: 600;
  font-size: 1rem;
  margin-bottom: 0.25rem;
  color: #fff;
}

.toast-message {
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.9rem;
  line-height: 1.4;
}

.toast-icon {
  margin-top: 0.125rem;
  font-size: 1.25rem;
}

.toast-close {
  margin-left: 0.5rem;
  color: rgba(255, 255, 255, 0.5) !important;
  min-width: 32px;
  height: 32px;
}

.toast-close:hover {
  color: rgba(255, 255, 255, 0.8) !important;
  background: rgba(255, 255, 255, 0.1) !important;
}

.toast.success { 
  border-left: 4px solid #10B981; 
  background: rgba(16, 185, 129, 0.05);
}

.toast.error { 
  border-left: 4px solid #EF4444; 
  background: rgba(239, 68, 68, 0.05);
}

.toast.warning { 
  border-left: 4px solid #F59E0B; 
  background: rgba(245, 158, 11, 0.05);
}

.toast.info { 
  border-left: 4px solid #60a5fa; 
  background: rgba(96, 165, 250, 0.05);
}

.toast-fade-slide-enter-active, .toast-fade-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.toast-fade-slide-enter-from {
  opacity: 0;
  transform: translateX(100%) translateY(-10px);
}

.toast-fade-slide-leave-to {
  opacity: 0;
  transform: translateX(100%) translateY(-10px);
}

.toast-fade-slide-move {
  transition: transform 0.3s ease;
}

@keyframes toast-fade-in {
  from { 
    opacity: 0; 
    transform: translateX(100%) translateY(-10px); 
  }
  to { 
    opacity: 1; 
    transform: translateX(0) translateY(0); 
  }
}

/* Responsive design */
@media (max-width: 768px) {
  .notification-container {
    top: 1rem;
    right: 1rem;
    left: 1rem;
  }
  
  .toast {
    min-width: auto;
    max-width: none;
    width: 100%;
  }
}
</style> 