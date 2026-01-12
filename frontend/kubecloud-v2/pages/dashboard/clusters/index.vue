<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h5 font-weight-bold">
        Kubernetes Clusters
      </h1>
      <p class="text-body-2 mt-1 text-accent">
        Manage your cloud-native infrastructure
      </p>
    </div>

    <div class="d-flex justify-end ga-4 mb-8">
      <v-btn prepend-icon="mdi-plus" text="New Cluster" variant="tonal" color="primary" />
      <v-btn
        prepend-icon="mdi-trash-can-outline"
        color="error"
        variant="text"
        text="Delete All"
        border
        class="border-error"
        :style="{ '--v-border-opacity': '0.3' }"
      />
    </div>

    <v-row class="mb-4">
      <v-col cols="9">
        <v-text-field
          label="Search by name"
          placeholder="Search by name"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
          density="compact"
          hide-details
        />
      </v-col>

      <v-col cols="3">
        <v-select
          label="Sort By"
          :items="sortByItems"
          variant="outlined"
          density="compact"
          hide-details
          clearable
        />
      </v-col>
    </v-row>

    <v-data-table
      :headers="[
        { title: 'Name', key: 'name', align: 'center', sortable: false },
        { title: 'Nodes', key: 'nodes', align: 'center', sortable: false },
        { title: 'Created At', key: 'createdAt', align: 'center', sortable: false },
        { title: 'Actions', key: 'actions', align: 'center', sortable: false },
      ]"
      :items="deployments"
      :loading="isLoading"
    >
      <template #item="{ item }">
        <DeploymentRow :deployment="item" />
      </template>
    </v-data-table>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state: deployments, isLoading } = useAsyncState(async () => {
  const { data } = await api.deployments.deploymentsGet()
  return data.data?.deployments ?? []
}, [])

const sortByItems = [
  { title: "Name", value: "name" },
  { title: "Created At", value: "createdAt" },
  { title: "Nodes", value: "nodes" },
]
</script>
