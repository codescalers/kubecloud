<template>
  <v-dialog v-model="dialog" max-width="400">
    <BaseDialogCard>
      <template #title>
        Delete Cluster
      </template>
      <template #content>
        <div class="delete-message">Are you sure you want to delete this cluster? This action cannot be undone.</div>
      </template>
      <template #actions>
        <v-btn variant="outlined" color="error" :loading="loading" @click="emit('confirm')">Delete</v-btn>
        <v-btn variant="outlined" @click="closeDialog">Cancel</v-btn>
      </template>
    </BaseDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import BaseDialogCard from './BaseDialogCard.vue';
import { computed } from 'vue';
const props = defineProps<{
  modelValue: boolean,
  loading: boolean
}>();
const emit = defineEmits(['update:modelValue', 'confirm']);
const dialog = computed({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val)
});
function closeDialog() { emit('update:modelValue', false); }
</script>

<style scoped>
.delete-message {
  color: #ff6b6b;
  font-weight: 500;
  margin-bottom: 1rem;
}
</style>
