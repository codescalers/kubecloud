<template>
  <div class="pa-4 pb-0">
    <div class="d-flex justify-space-between align-center">
      <p v-text="item.name" class="text-body-1 mb-0 font-weight-bold" />
      <p
        v-text="item.status === 'running' ? 'In Progress' : item.status"
        class="ml-8 text-capitalize mb-0 text-gray-500"
        :class="{
          'text-success': item.status === 'success',
          'text-error': item.status === 'failed',
          'text-primary': item.status === 'running',
        }"
      />
    </div>
    <div class="d-flex justify-space-between align-center text-sm opacity-80">
      <p v-text="moment(item.created_at).fromNow()" class="mb-0" />
      <p
        v-text="item.current_step + '/' + item.total_steps"
        class="mb-0"
        v-if="item.status !== 'pending'"
      />
      <p v-text="'N/A'" class="mb-0" v-else />
    </div>
    <v-progress-linear
      height="2"
      :model-value="item.status === 'pending' ? 0 : (item.current_step / item.total_steps) * 100"
      :color="
        item.status === 'running'
          ? 'primary'
          : item.status === 'success'
            ? 'success'
            : item.status === 'failed'
              ? 'error'
              : undefined
      "
      class="mt-2 mb-2"
    />
  </div>
</template>

<script setup lang="ts">
import moment from 'moment'
import type { UserWorkflowsResponse } from '@/utils/userService'

defineProps<{ item: UserWorkflowsResponse }>()
</script>
