<template>
  <TransitionGroup 
    name="notification" 
    tag="div" 
    class="notification-container"
  >
    <div
      v-for="notification in notifications"
      :key="notification.id"
      class="notification-toast"
      :class="notification.type"
    >
      <div class="notification-icon">
        <v-icon :icon="getIcon(notification.type)" size="24"></v-icon>
      </div>
      <div class="notification-content">
        <h4 class="notification-title">{{ notification.title }}</h4>
        <p class="notification-message">{{ notification.message }}</p>
      </div>
      <div class="notification-actions">
        <v-btn
          v-if="notification.action"
          variant="text"
          size="small"
          @click="handleAction(notification)"
          class="action-btn"
        >
          {{ notification.action.text }}
        </v-btn>
        <v-btn
          icon
          variant="text"
          size="small"
          @click="removeNotification(notification.id)"
          class="close-btn"
        >
          <v-icon icon="mdi-close" size="16"></v-icon>
        </v-btn>
      </div>
      <div class="notification-progress">
        <div 
          class="progress-bar"
          :style="{ animationDuration: `${notification.duration}ms` }"
        ></div>
      </div>
    </div>
  </TransitionGroup>
</template>

<script setup lang="ts">
import { useNotificationStore } from '../stores/notifications'

const notificationStore = useNotificationStore()
const { notifications, removeNotification } = notificationStore

const getIcon = (type: string) => {
  const icons = {
    success: 'mdi-check-circle',
    error: 'mdi-alert-circle',
    warning: 'mdi-alert',
    info: 'mdi-information'
  }
  return icons[type as keyof typeof icons] || 'mdi-information'
}

const handleAction = (notification: any) => {
  if (notification.action) {
    notification.action.handler()
  }
  removeNotification(notification.id)
}
</script>

<style scoped>
.notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 400px;
}

.notification-toast {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem;
  border-radius: 12px;
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 2px solid;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  position: relative;
  overflow: hidden;
  min-width: 300px;
}

.notification-toast.success {
  border-color: var(--kubecloud-blue);
  box-shadow: 0 8px 32px rgba(37, 99, 235, 0.3);
}

.notification-toast.error {
  border-color: var(--neon-rose);
  box-shadow: 0 8px 32px rgba(239, 68, 68, 0.3);
}

.notification-toast.warning {
  border-color: var(--kubecloud-orange);
  box-shadow: 0 8px 32px rgba(249, 115, 22, 0.3);
}

.notification-toast.info {
  border-color: var(--neon-cyan);
  box-shadow: 0 8px 32px rgba(6, 182, 212, 0.3);
}

.notification-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.notification-toast.success .notification-icon {
  color: var(--kubecloud-blue);
}

.notification-toast.error .notification-icon {
  color: var(--neon-rose);
}

.notification-toast.warning .notification-icon {
  color: var(--kubecloud-orange);
}

.notification-toast.info .notification-icon {
  color: var(--neon-cyan);
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.25rem 0;
  color: var(--kubecloud-white);
}

.notification-message {
  font-size: 0.9rem;
  color: var(--kubecloud-light-gray);
  margin: 0;
  line-height: 1.4;
}

.notification-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-end;
}

.action-btn {
  font-size: 0.8rem;
  padding: 0.25rem 0.5rem;
  min-width: auto;
  height: auto;
}

.close-btn {
  opacity: 0.7;
  transition: opacity 0.2s;
}

.close-btn:hover {
  opacity: 1;
}

.notification-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: rgba(255, 255, 255, 0.1);
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--kubecloud-blue), var(--kubecloud-orange));
  animation: progress-shrink linear forwards;
}

@keyframes progress-shrink {
  from { width: 100%; }
  to { width: 0%; }
}

/* Transition animations */
.notification-enter-active,
.notification-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.notification-enter-from {
  opacity: 0;
  transform: translateX(100%) scale(0.9);
}

.notification-leave-to {
  opacity: 0;
  transform: translateX(100%) scale(0.9);
}

.notification-move {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@media (max-width: 600px) {
  .notification-container {
    top: 10px;
    right: 10px;
    left: 10px;
    max-width: none;
  }
  
  .notification-toast {
    min-width: auto;
  }
}
</style> 