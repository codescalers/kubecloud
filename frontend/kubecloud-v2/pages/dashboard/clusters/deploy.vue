<template>
  <v-container class="py-16">
    <div class="mb-16">
      <h1 class="text-h2 font-weight-bold text-center mb-4">
        Deploy New Cluster
      </h1>
      <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
        Create and configure your Kubernetes cluster in just a few steps
      </p>
    </div>

    {{ cluster }}

    <v-stepper
      v-model="step"
      flat
      bg-color="transparent"
      #="{ next, prev }"
    >
      <v-stepper-header class="elevation-0">
        <v-stepper-item :value="1" title="Define VMs" />
        <v-divider />
        <v-stepper-item :value="2" title="Place VMs" />
        <v-divider />
        <v-stepper-item :value="3" title="Review" />
      </v-stepper-header>

      <v-stepper-window>
        <v-stepper-window-item :value="1">
          <DefineVMsForm v-model="cluster" />
        </v-stepper-window-item>

        <v-stepper-window-item :value="2">
          Place VMs
        </v-stepper-window-item>

        <v-stepper-window-item :value="3">
          Review
        </v-stepper-window-item>
      </v-stepper-window>

      <v-divider class="mt-12 mb-6" />

      <div class="d-flex justify-end ">
        <v-btn
          prepend-icon="mdi-arrow-left"
          text="Back"
          variant="plain"
          :disabled="step === 1"
          @click="prev"
        />

        <v-btn
          append-icon="mdi-arrow-right"
          text="Next"
          variant="tonal"
          color="primary"
          @click="next"
        />
      </div>
    </v-stepper>
  </v-container>
</template>

<script setup lang="ts">
const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onUnmounted(drawer.open)

onMounted(container.fluidize)
onUnmounted(container.containerize)

const step = ref(1)

const cluster = ref<ClusterForm>({
  name: "engine789",
  masters: [createClusterNode({ name: "Leader", permanent: true })],
  workers: [],
})
</script>
