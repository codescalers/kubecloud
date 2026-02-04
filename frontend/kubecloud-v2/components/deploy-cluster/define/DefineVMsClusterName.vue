<template>
  <v-card>
    <h3 class="text-h6 font-weight-bold mb-6">
      Cluster Configuration
    </h3>

    <v-text-field
      :model-value="$props.modelValue"
      variant="outlined"
      label="Cluster Name"
      placeholder="Enter cluster name"
      hint="Cluster name must be at least 3 characters and contain only letters and numbers"
      persistent-hint
      :rules="[
        (v) => !!v || 'Cluster name is required', (v) => v.length >= 3 || 'Cluster name must be at least 3 characters',
        (v) => v.length <= 255 || 'Cluster name must be less than 255 characters',
        (v) => /^[a-zA-Z0-9]+$/.test(v) || 'Cluster name must contain only letters and numbers',
      ]"
      @update:model-value="$emit('update:modelValue', $event)"
    >
      <template #append-inner>
        <v-icon icon="mdi-refresh" tabindex="-1" @click="console.log($event)" />
      </template>
    </v-text-field>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ modelValue: string }>()
defineEmits<{ (e: "update:modelValue", value: string): void }>()
</script>
