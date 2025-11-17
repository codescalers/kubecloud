<template>
  <v-menu location="bottom end">
    <template v-slot:activator="{ props }">
      <v-btn icon variant="text" color="white" v-bind="props" :disabled="!active">
        <div :style="{ position: 'relative' }">
          <v-icon
            icon="mdi-lightning-bolt"
            size="24"
            :color="active ? 'primary' : undefined"
            :style="{
              position: 'absolute',
              top: '0',
              left: '0',
              transform: 'translate(-50%, -50%)',
            }"
          />
          <v-icon
            v-if="active"
            icon="mdi-lightning-bolt"
            size="24"
            color="primary"
            :style="{
              position: 'absolute',
              top: '0',
              left: '0',
              transform: 'translate(-50%, -50%)',
              animation: 'pulse 1s infinite',
            }"
          />
        </div>
      </v-btn>
    </template>
    <v-list>
      <v-list-item>
        <v-list-item-title>
          <v-icon icon="mdi-delete" size="24"></v-icon>
          Delete
        </v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const active = ref(true)
</script>

<style>
@keyframes pulse {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(2);
  }
}
</style>
