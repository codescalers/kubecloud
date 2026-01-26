<template>
  <div ref="vantaDots">
    <div
      class="d-flex justify-center align-center"
      :style="{
        minHeight: '700px',
        background: `
            url(${myceliumBg1}) no-repeat  bottom right/40%,
            url(${myceliumBg2}) no-repeat  -120px 80px/30%
          `,
      }"
    >
      <v-container>
        <div>
          <h1 class="text-h1 font-weight-bold text-center mb-4">
            Mycelium Cloud
          </h1>
          <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
            Revolutionary Kubernetes platform that transforms how teams deploy and manage cloud-native
            applications at scale
          </p>
        </div>
      </v-container>
    </div>
  </div>

  <VDivider />
</template>

<script setup lang="ts">
import myceliumBg1 from "~/assets/images/mycelium_bg_1.svg"
import myceliumBg2 from "~/assets/images/mycelium_bg_2.svg"

ensureThreeGlobal()

const vantaDots = ref<HTMLElement | null>(null)

useHead({
  script: [
    {
      async: true,
      src: "/scripts/vanta.dots.min.js",
      defer: true,
      onload() {
        window?.VANTA?.DOTS?.({
          el: vantaDots.value,
          mouseControls: true,
          touchControls: true,
          gyroControls: false,
          minHeight: 200,
          minWidth: 200,
          scale: 1,
          scaleMobile: 1,
          color: 0x2B3951,
          color2: 0x2B3951,
          backgroundColor: 0x0A192F,
          size: 5,
          spacing: 70,
          showLines: true,
        })
      },
    },
  ],
})
</script>
