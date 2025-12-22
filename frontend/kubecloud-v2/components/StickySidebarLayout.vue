<template>
  <v-container>
    <v-row>
      <v-col :cols="12 - cols">
        <v-list class="sticky-top">
          <slot name="items" />
        </v-list>
      </v-col>

      <v-col :cols="cols">
        <slot />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
defineProps<{ cols: number }>()
</script>
