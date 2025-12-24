<template>
  <v-row no-gutters class="h-100">
    <v-col class="mh-100 overflow-hidden" cols="12" md="6">
      <div class="h-100 d-flex align-center">
        <slot />
      </div>
    </v-col>

    <v-col cols="6" class="d-none d-md-block" :order="reversed ? 'first' : 'last'">
      <div
        class="h-100 w-100 bg-primary d-flex justify-center align-center px-10"
        :style="{
          background: `url(${reversed ? myceliumBg2 : myceliumBg1}) no-repeat  ${reversed ? '-250px 170px' : 'bottom right'}/70%, linear-gradient(180deg, #182F66 0%, #0F172A 100%)`,
        }"
      >
        <div :style="{ maxWidth: '600px' }">
          <h2 class="text-h4 font-weight-bold" v-text="title" />
          <div class="text-h6">
            <p class="mt-3 mb-8" v-text="quote" />

            <p v-text="subtitle" />
          </div>
        </div>
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import myceliumBg1 from "~/assets/images/mycelium_bg_1.svg"
import myceliumBg2 from "~/assets/images/mycelium_bg_2.svg"

defineProps<{
  title: string
  quote: string
  subtitle: string
  reversed?: boolean
}>()
</script>
