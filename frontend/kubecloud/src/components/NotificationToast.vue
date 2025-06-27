<template>
  <transition name="toast-fade-slide">
    <div v-if="visible" :class="['toast', status]">
      <v-icon class="toast-icon" :color="iconColor" left>{{ icon }}</v-icon>
      <span class="toast-message">{{ message }}</span>
      <v-btn icon class="toast-close" @click="close"><v-icon>mdi-close</v-icon></v-btn>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
const props = defineProps<{ message: string, status: 'success' | 'error' | 'info', duration?: number }>()
const visible = ref(true)
const icon = props.status === 'success' ? 'mdi-check-circle' : props.status === 'error' ? 'mdi-alert-circle' : 'mdi-information'
const iconColor = props.status === 'success' ? '#10B981' : props.status === 'error' ? '#EF4444' : '#60a5fa'
function close() { visible.value = false }
onMounted(() => { setTimeout(close, props.duration || 3500) })
watch(() => props.message, () => { visible.value = true })
</script>

<style scoped>
.toast {
  display: flex;
  align-items: center;
  background: #202632;
  color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(96, 165, 250, 0.13);
  padding: 1rem 1.5rem;
  min-width: 280px;
  max-width: 400px;
  font-weight: 600;
  font-size: 1rem;
  position: fixed;
  top: 2rem;
  right: 2rem;
  z-index: 9999;
  border: 1px solid #232A36;
  animation: toast-fade-in 0.3s;
}
.toast-icon {
  margin-right: 1rem;
  font-size: 1.5rem;
}
.toast-close {
  margin-left: 1rem;
  color: #7b8494 !important;
}
.toast.success { border-left: 4px solid #10B981; }
.toast.error { border-left: 4px solid #EF4444; }
.toast.info { border-left: 4px solid #60a5fa; }
.toast-fade-slide-enter-active, .toast-fade-slide-leave-active {
  transition: opacity 0.3s, transform 0.3s;
}
.toast-fade-slide-enter-from, .toast-fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}
@keyframes toast-fade-in {
  from { opacity: 0; transform: translateY(-20px); }
  to { opacity: 1; transform: translateY(0); }
}
</style> 