<template>
  <div class="border rounded-xl pa-4 border-dashed">
    <div class="d-flex align-center ga-2">
      <h5 class="text-h6 font-weight-bold" v-text="modelValue.name" />
      <v-chip color="primary" size="x-small" :text="modelValue.type" class="rounded font-weight-bold text-caption text-capitalize" />
    </div>

    <v-chip
      v-if="modelValue.useFullNodeCapabilities"
      color="success"
      size="small"
      prepend-icon="mdi-check-decagram"
      class="font-weight-bold mt-2"
    >
      Use Full Node Capabilities
    </v-chip>

    <div v-else class="d-flex align-center flex-wrap ga-2 mt-2">
      <v-chip size="small" prepend-icon="mdi-cpu-64-bit" color="primary" class="font-weight-bold">
        CPU: {{ modelValue.cpu }} vCores
      </v-chip>
      <v-chip size="small" prepend-icon="mdi-memory" color="success" class="font-weight-bold">
        RAM: {{ modelValue.memory }} GB
      </v-chip>
      <v-chip size="small" prepend-icon="mdi-server" color="secondary" class="font-weight-bold">
        Disk Size: {{ modelValue.disk }} GB
      </v-chip>
    </div>

    <v-select
      class="mt-8"
      label="Select Node"
      placeholder="pick a node"
      :items="['node 1', 'node 2']"
      density="compact"
      variant="outlined"
    />
  </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: ClusterNode }>()
</script>
