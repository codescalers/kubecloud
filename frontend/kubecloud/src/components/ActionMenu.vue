<template>
  <v-menu location="bottom end" max-width="400px">
    <template v-slot:activator="{ props }">
      <v-btn icon variant="text" color="white" v-bind="props" :disabled="!active">
        <div :style="{ position: 'relative' }">
          <v-icon
            icon="mdi-lightning-bolt"
            size="24"
            :color="active ? 'primary' : undefined"
            :style="{
              position: 'absolute',
              top: '0',
              left: '0',
              transform: 'translate(-50%, -50%)',
            }"
          />
          <v-icon
            v-if="active"
            icon="mdi-lightning-bolt"
            size="24"
            color="primary"
            :style="{
              position: 'absolute',
              top: '0',
              left: '0',
              transform: 'translate(-50%, -50%)',
              animation: 'pulse 1s infinite',
            }"
          />
        </div>
      </v-btn>
    </template>
    <v-list>
      <template v-for="(item, index) in items" :key="item.id">
        <v-divider v-if="index > 0" />
        <ActionMenuItem :item="item" />
      </template>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ActionMenuItem, { type ActionItem } from './ActionMenuItem.vue'

const active = ref(true)
const items = ref<ActionItem[]>([
  {
    id: '1',
    title: 'Add Node to Cluster',
    status: 'pending',
    createdAt: 'Created 5m ago',
    currentStep: 0,
    totalSteps: 3,
  },
  {
    id: '2',
    title: 'Create Cluster',
    status: 'running',
    createdAt: 'Running 2m ago',
    currentStep: 2,
    totalSteps: 4,
  },
  {
    id: '3',
    title: 'Process Payment',
    status: 'failed',
    createdAt: 'Completed 12m ago',
    currentStep: 3,
    totalSteps: 5,
  },
  {
    id: '3',
    title: 'Create Cluster',
    status: 'success',
    createdAt: 'Completed 2d ago',
    currentStep: 4,
    totalSteps: 4,
  },
])
</script>

<style>
@keyframes pulse {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(2.5);
  }
}
</style>
