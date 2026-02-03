<template>
  <v-card :style="{ padding: '0 !important' }">
    <v-card-title class="pa-8 d-flex align-center justify-space-between border-b">
      <div>
        <div class="d-flex ga-2 align-baseline">
          <v-icon icon="mdi-sitemap-outline" size="small" color="success" />
          <span class="text-h5 font-weight-bold">Workflow Details</span>
        </div>
        <p class="text-subtitle-2 opacity-50">
          Review execution parameters and current state
        </p>
      </div>

      <v-btn
        icon
        size="small"
        variant="plain"
        :style="{ borderRadius: '50% !important' }"
        @click="$emit('cancel')"
      >
        <v-icon icon="mdi-close" size="large" @click="$emit('cancel')" />
      </v-btn>
    </v-card-title>

    <v-card-text>
      <v-row>
        <v-col cols="12">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            uuid
          </p>
          <p
            class="text-subtitle-2 border d-inline-block px-3 py-1 rounded-lg"
            :style="{ backgroundColor: 'rgb(var(--v-bg-2))' }"
          >
            <span class="opacity-70" v-text="workflow?.uuid" />
          </p>
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            name
          </p>
          <p class="text-subtitle-2" v-text="workflow?.display_name" />
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            template name
          </p>
          <div class="d-flex align-center ga-2 opacity-70">
            <v-icon icon="mdi-puzzle" size="x-small" />
            <p class="text-subtitle-2 border px-2 rounded-lg" v-text="workflow?.name" />
          </div>
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            Status
          </p>
          <div class="d-inline-block">
            <StatusChip :status="workflow?.status!" />
          </div>
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            Progress
          </p>
          <div class="text-subtitle-2 mb-2 d-flex align-center ga-3">
            <span>
              Step {{ workflow?.current_step! }} of
              {{ workflow?.total_steps! }}
            </span>
            <v-avatar size="6" color="white" class="opacity-50" />
            <span class="opacity-50">{{ workflow?.step_name! }}</span>
          </div>
          <v-progress-linear
            rounded
            :model-value="(workflow?.current_step! * 100) / workflow?.total_steps!"
            :color="color"
            :chunk-count="workflow?.total_steps!"
            chunk-gap="5"
          />
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            user id
          </p>
          <p
            class="text-subtitle-2"
            :class="{ 'opacity-50': !workflow?.user_id }"
            v-text="workflow?.user_id || 'N/A'"
          />
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            queue
          </p>
          <p
            class="text-subtitle-2"
            :class="{ 'opacity-50': !workflow?.queue_name }"
            v-text="workflow?.queue_name || 'N/A'"
          />
        </v-col>

        <v-col cols="6">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            created at
          </p>
          <p class="text-subtitle-2 d-flex align-center ga-2">
            <v-icon icon="mdi-calendar-outline" size="x-small" />
            <span>{{ createdAt }}</span>
          </p>
        </v-col>

        <v-col v-if="workflow?.metadata && Object.keys(workflow.metadata).length > 0" cols="12">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            metadata
          </p>
          <div :style="{ margin: '-20px 0 -24px 0' }" v-html="metadata" />
        </v-col>

        <v-col cols="12">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            state
          </p>
          <div :style="{ margin: '-20px 0 -24px 0' }" v-html="state" />
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { ServicesAdminWorkflow } from "../generated/api"

const props = defineProps<{ workflow?: ServicesAdminWorkflow }>()
defineEmits<{ (e: "cancel"): void }>()

const color = useStatusColor(() => props.workflow?.status ?? "")
const createdAt = useDateFormat(() => props.workflow?.created_at, "DD/MM/YYYY, HH:mm")
const metadata = computed(() => marked.parse(`\`\`\`json\n${JSON.stringify(props.workflow?.metadata, null, 2)}\n\`\`\``, { renderer }))
const state = computed(() => marked.parse(`\`\`\`json\n${JSON.stringify(props.workflow?.state, null, 2)}\n\`\`\``, { renderer }))
</script>
