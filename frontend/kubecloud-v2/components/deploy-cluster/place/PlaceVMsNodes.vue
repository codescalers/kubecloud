<template>
  <v-card>
    <div class="d-flex align-center justify-space-between mb-8">
      <h3 class="d-flex align-center ga-2">
        <v-icon icon="mdi-sitemap-outline" size="x-small" color="primary" />
        Cluster Nodes
      </h3>
    </div>

    <v-row>
      <v-col v-for="(node, idx) in cluster.masters" :key="node.id" cols="6">
        <PlaceVMsNode v-model="$props.cluster.masters[idx]!" />
      </v-col>

      <v-col v-for="(node, idx) in cluster.workers" :key="node.id" cols="6">
        <PlaceVMsNode v-model="$props.cluster.workers[idx]!" />
      </v-col>
    </v-row>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ cluster: ClusterForm }>()
</script>
