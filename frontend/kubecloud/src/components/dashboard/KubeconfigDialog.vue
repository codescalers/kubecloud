<template>
  <v-dialog v-model="dialog" max-width="600">
    <BaseDialogCard>
      <template #title>
        Kubeconfig
      </template>
      <template #content>
        <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />
        <v-alert v-if="error" type="error" class="mb-4">{{ error }}</v-alert>
        <div v-if="content && !loading && !error">
          <pre>{{ content }}</pre>
        </div>
        <div v-else-if="!loading && !error" class="empty-message">No kubeconfig available.</div>
      </template>
      <template #actions>
        <v-btn variant="outlined" color="primary" @click="emit('copy')">Copy</v-btn>
        <v-btn variant="outlined" color="primary" @click="emit('download')">Download</v-btn>
      </template>
    </BaseDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import BaseDialogCard from './BaseDialogCard.vue';
const props = defineProps<{
  modelValue: boolean,
  projectName: string,
  loading: boolean,
  error: string,
  content: string
}>();
const emit = defineEmits(['update:modelValue', 'copy', 'download']);
const dialog = computed({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val)
});
function closeDialog() { emit('update:modelValue', false); }
</script>

<style scoped>
pre {
  color: #b4befe;
  font-family: 'JetBrains Mono', 'Fira Mono', 'Menlo', 'Consolas', monospace;
  font-size: 1rem;
  line-height: 1.6;
  padding: 1.5rem;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  overflow-x: auto;
  overflow-y: auto;
  max-height: 60vh;
  box-sizing: border-box;
}
.empty-message {
  color: #888;
  font-style: italic;
  margin-bottom: 0.5rem;
}
</style>
