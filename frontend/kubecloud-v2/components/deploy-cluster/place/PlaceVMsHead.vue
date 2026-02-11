<template>
  <v-card>
    <h3 class="text-h6 font-weight-bold">
      Assign VMs to Nodes
    </h3>
    <p class="text-caption text-accent mb-6">
      Select nodes to host your cluster VMs
    </p>

    <v-alert
      type="info"
      variant="tonal"
      title="50% Discount Available!"
      text="Reserve a node to get 50% discount and exclusive usage. Shared nodes are available at full price."
      class="mb-8"
    >
      <div class="d-flex justify-end">
        <v-btn
          class="mt-2"
          text="Reserve Node"
          append-icon="mdi-arrow-right"
          variant="plain"
          color="white"
          border
          :to="ROUTES.Dashboard.MyNodes.Explorer()"
        />
      </div>
    </v-alert>

    <v-select
      v-model="$props.modelValue.region"
      label="Filter By Region"
      placeholder="Pick a region"
      clearable
      :items="['Europe', 'Asia', 'America', 'Africa', 'Australia']"
      variant="outlined"
      density="compact"
      hide-details
      @update:model-value="() => {
        $props.modelValue.masters = $props.modelValue.masters.map(master => ({ ...master, node: null }))
        $props.modelValue.workers = $props.modelValue.workers.map(worker => ({ ...worker, node: null }))
      }"
    />
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ modelValue: ClusterForm }>()
</script>
