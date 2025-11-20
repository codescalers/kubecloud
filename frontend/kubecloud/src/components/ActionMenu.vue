<template>
  <v-menu
    location="bottom end"
    max-width="400px"
    :model-value="openMenu && active"
    @update:model-value="openMenu = $event && active"
  >
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
      <template v-for="(item, index) in workflows" :key="item.workflow_id">
        <v-divider v-if="index > 0" />
        <ActionMenuItem :item="item" />
      </template>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, onBeforeUnmount } from 'vue'
import ActionMenuItem from './ActionMenuItem.vue'
import { userService, type UserWorkflowsResponse } from '@/utils/userService'

const workflows = ref<UserWorkflowsResponse[]>([])
const openMenu = ref(false)
const active = computed(() => workflows.value.length > 0)

let intervalId: NodeJS.Timeout | null = null
onBeforeUnmount(() => intervalId && clearInterval(intervalId))
onMounted(() => {
  intervalId = setInterval(loadWorkflows, 3000)
  loadWorkflows()
})

async function loadWorkflows() {
  try {
    workflows.value = await userService.getWorkflows()
  } catch (error) {
    console.log({ error })
  }
}
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
