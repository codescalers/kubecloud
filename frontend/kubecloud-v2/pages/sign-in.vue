<template>
  <div
    class="bg-red h-100"
    :style="{
      background: `
        url(${myceliumBg1}) no-repeat  bottom right/40%,
        url(${myceliumBg2}) no-repeat  -200px 170px/30%,
        linear-gradient(100.95deg, #0B1A30 71.46%, #225196 145.62%)
      `,
    }"
  >
    <div class="form-container">
      <v-card max-width="461px" class="w-100">
        <SignInForm />
      </v-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import myceliumBg1 from "~/assets/images/mycelium_bg_1.svg"
import myceliumBg2 from "~/assets/images/mycelium_bg_2.svg"

definePageMeta({ middleware: "non-auth" })
</script>
