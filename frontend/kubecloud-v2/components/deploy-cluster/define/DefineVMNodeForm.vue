<template>
  <v-row>
    <v-col cols="12" class="pb-0">
      <v-text-field v-model="$props.node.name" variant="outlined" label="Node Name" />
    </v-col>

    <v-col cols="12">
      <div class="d-flex justify-space-between align-center">
        <div>
          <p class="text-subtitle-1 font-weight-bold">
            Use Full Node Capabilities
          </p>
          <p class="text-caption opacity-70">
            Include all available resources within placed node
          </p>
        </div>

        <v-switch v-model="$props.node.useFullNodeCapabilities" color="primary" inset hide-details />
      </div>
    </v-col>

    <v-expand-transition>
      <div v-if="!node.useFullNodeCapabilities" class="w-100">
        <v-col cols="12" class="pb-0">
          <v-row>
            <v-col cols="4" class="pb-0">
              <v-text-field v-model="$props.node.cpu" variant="outlined" label="CPU (vCores)" />
            </v-col>

            <v-col cols="4" class="pb-0">
              <v-text-field v-model="$props.node.memory" variant="outlined" label="RAM (GB)" />
            </v-col>

            <v-col cols="4" class="pb-0">
              <v-text-field v-model="$props.node.disk" variant="outlined" label="Disk Size (GB)" />
            </v-col>
          </v-row>
        </v-col>
      </div>
    </v-expand-transition>

    <v-col cols="12" class="pb-0 d-flex align-start ga-3">
      <label class="text-body-2 text-no-wrap mt-2">
        SSH Keys:
      </label>

      <v-chip-group v-model="$props.node.sshKeys" multiple selected-class="text-primary">
        <v-chip
          v-for="(key, index) in sshKeys"
          :key="key.id"
          filter
          :value="index"
          class="rounded"
          size="small"
        >
          {{ key.name }}
        </v-chip>
      </v-chip-group>
    </v-col>
    <p v-if="node.sshKeys.length === 0" class="text-error text-caption">
      <v-icon icon="mdi-alert-circle" size="small" color="error" />
      Please select at least one SSH key
    </p>
  </v-row>
</template>

<script setup lang="ts">
import type { ModelsSSHKey } from "~/generated/api"

defineProps<{ sshKeys: ModelsSSHKey[], node: ClusterNode }>()
</script>
