<template>
  <tr class="text-no-wrap py-2">
    <td>
      <span class="text-subtitle-2" v-text="workflow.display_name" />
      <span v-if="workflow.name" class="d-block opacity-50 text-caption">
        ({{ workflow.name }})
      </span>
    </td>

    <td>
      <StatusChip :status="workflow.status!" />
    </td>

    <td>
      <span class="text-caption opacity-50 d-block mb-1 font-weight-medium">
        {{ workflow.current_step }}/{{ workflow.total_steps }}
      </span>
      <div :style="{ width: '100px' }">
        <v-progress-linear
          rounded
          :model-value="(workflow.current_step! * 100) / workflow.total_steps!"
          :color="color"
          :chunk-count="workflow.total_steps!"
          chunk-gap="5"
        />
      </div>
    </td>

    <td class="text-subtitle-2 opacity-50 text-center">{{ workflow.step_name }}</td>
    <td class="text-subtitle-2 opacity-50 text-center">{{ workflow.user_id || "-" }}</td>
    <td class="text-subtitle-2 opacity-50 text-center">
      {{ createdAt }}
    </td>
    <td>
      <v-btn
        variant="plain"
        class="border"
        prepend-icon="mdi-eye-outline"
        size="small"
        text="View"
        @click="$emit('view')"
      />
    </td>
  </tr>
</template>

<script setup lang="ts">
import type { ServicesAdminWorkflow } from "../generated/api"

const props = defineProps<{ workflow: ServicesAdminWorkflow }>()
defineEmits<{ (e: "view"): void }>()

const color = useStatusColor(() => props.workflow.status!)
const createdAt = useDateFormat(() => props.workflow.created_at, "DD/MM/YYYY, HH:mm")
</script>
