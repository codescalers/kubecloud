<template>
  <div ref="vantaGlobe">
    <div
      class="d-flex justify-center align-center"
      :style="{
        minHeight: '600px',

      }"
    >
      <v-container class="mt-16">
        <div>
          <h1 class="text-h1 font-weight-bold text-center mb-8">
            Ready to Transform Your Kubernetes Experience?
          </h1>
          <p class="text-body-1 text-center mx-auto pa-2 rounded-lg" :style="{ maxWidth: '600px', backgroundColor: 'rgba(var(--v-theme-background), 0.8)' }">
            Join a growing community of developers and DevOps engineers who trust Mycelium Cloud for their production workloads.
          </p>

          <div class="d-flex justify-center mt-4">
            <v-btn
              class="btn-form"
              :style="{ color: '#333 !important' }"
              text="Deploy Kubernetes Cluster"
              variant="flat"
              color="primary"
              to="/dashboard/clusters/deploy"
            />
          </div>
        </div>
      </v-container>
    </div>
  </div>
</template>

<script setup lang="ts">
ensureThreeGlobal()

const vantaGlobe = ref<HTMLElement | null>(null)
useVanta("globe", () => vantaGlobe.value!)
</script>
