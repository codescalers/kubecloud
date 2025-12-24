<template>
  <v-row no-gutters class="h-100">
    <!-- class="px-10 bg-surface" -->
    <v-col
      cols="12"
      md="6"
      :style="{
        maxHeight: '100%',
        overflow: 'hidden',
      }"
    >
      <div class="h-100 d-flex align-center">
        <slot />
      </div>
    </v-col>

    <v-col cols="6" class="d-none d-md-block">
      <div
        class="h-100 w-100 bg-primary d-flex justify-center align-center px-10"
        :style="{
          background: `url(${myceliumBg1}) no-repeat bottom right/70%, linear-gradient(180deg, #182F66 0%, #0F172A 100%)`,
        }"
      >
        <div :style="{ maxWidth: '560px' }">
          <h2 :style="{ fontSize: '32px', fontWeight: '700' }">Join Mycelium Cloud</h2>
          <div :style="{ fontSize: '20px' }">
            <p class="mt-3 mb-8">Start your journey into decentralized cloud infrastructure.</p>

            <p>
              Create your account to access secure compute, scalable resources, and a
              next-generation cloud built without central points of failure.
            </p>
          </div>
        </div>
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
// import { useDisplay } from "vuetify"
import myceliumBg1 from "~/assets/images/mycelium_bg_1.svg"

defineProps<{ cols: number; prependWorkart?: boolean; hideWorkart?: boolean }>()

// const { smAndDown } = useDisplay()
</script>
