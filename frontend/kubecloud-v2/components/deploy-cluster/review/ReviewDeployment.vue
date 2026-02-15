<template>
  <v-card>
    <h3 class="text-h6 font-weight-bold mb-6">
      Cluster Configuration
    </h3>

    <v-text-field
      :model-value="cluster.name"
      variant="outlined"
      label="Cluster Name"
      readonly
    />

    <div class="border border-dashed rounded-lg pa-6 position-relative">
      <h4
        class="text-body-2 font-weight-bold position-absolute bg-surface px-2"
        :style="{ top: 0, left: '16px', transform: 'translateY(-50%)' }"
      >
        Node Assignment
      </h4>

      <v-row>
        <v-col v-for="node in [...cluster.masters, ...cluster.workers]" :key="node.id" cols="6">
          <v-card rounded="lg" flat :style="{ padding: '16px !important' }">
            <div class="d-flex align-center ga-2">
              <h4 class="text-body-1 font-weight-bold" v-text="node.name" />
              <v-chip
                :text="node.type"
                size="small"
                rounded="lg"
                color="primary"
                class="text-capitalize font-weight-bold text-caption"
              />
            </div>

            <v-chip
              v-if="node.useFullNodeCapabilities"
              color="success"
              size="small"
              prepend-icon="mdi-check-decagram"
              class="font-weight-bold mt-2"
            >
              Use Full Node Capabilities
            </v-chip>

            <div v-else class="d-flex align-center flex-wrap ga-2 mt-2">
              <v-chip size="small" prepend-icon="mdi-cpu-64-bit" color="primary" class="font-weight-bold">
                CPU: {{ node.cpu }} vCores
              </v-chip>
              <v-chip size="small" prepend-icon="mdi-memory" color="success" class="font-weight-bold">
                RAM: {{ node.memory }} GiB
              </v-chip>
              <v-chip size="small" prepend-icon="mdi-server" color="secondary" class="font-weight-bold">
                Disk Size: {{ node.disk }} GiB
              </v-chip>
            </div>

            <div cols="12" class="pb-0 d-flex align-start ga-3">
              <label class="text-body-2 text-no-wrap mt-2">
                SSH Keys:
              </label>

              <v-chip-group :model-value="node.sshKeys" multiple selected-class="text-primary">
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
            </div>
          </v-card>
        </v-col>
      </v-row>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import type { ModelsSSHKey } from "~/generated/api"

defineProps<{ sshKeys: ModelsSSHKey[], cluster: ClusterForm }>()
</script>
