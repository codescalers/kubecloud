<template>
  <DefineTemplate>
    <v-col :cols="12 - cols">
      <div class="h-100 w-100 bg-primary d-flex justify-center align-center">KubeCloud</div>
    </v-col>
  </DefineTemplate>

  <v-row no-gutters class="h-100">
    <UseTemplate v-if="prependWorkart && !hideWorkart" />

    <v-col
      :cols="cols"
      class="px-10 bg-surface"
      :style="{
        maxHeight: '100%',
        overflow: 'hidden',
      }"
    >
      <div class="h-100 d-flex align-center">
        <slot />
      </div>
    </v-col>

    <UseTemplate v-if="!prependWorkart && !hideWorkart" />
  </v-row>
</template>

<script setup lang="ts">
defineProps<{ cols: number; prependWorkart?: boolean; hideWorkart?: boolean }>()

const [DefineTemplate, UseTemplate] = createReusableTemplate()
</script>
