<template>
  <div class="dashboard-card ssh-keys-card">
    <h2 class="dashboard-card-title mb-6">SSH Keys</h2>
    <div class="ssh-keys-list mb-6">
      <div v-for="key in sshKeys" :key="key.id" class="ssh-key-item">
        <div class="ssh-key-content">
          <div class="ssh-key-info">
            <div class="ssh-key-icon">
              <v-icon size="32" color="accent">mdi-key</v-icon>
            </div>
            <div class="ssh-key-details">
              <h3 class="ssh-key-name">{{ key.name }}</h3>
              <p class="ssh-key-fingerprint text-muted">{{ key.fingerprint }}</p>
              <p class="ssh-key-date text-muted">Added {{ key.addedDate }}</p>
            </div>
          </div>
          <div class="ssh-key-actions">
            <v-btn class="action-btn">Edit</v-btn>
            <v-btn class="action-btn">Remove</v-btn>
          </div>
        </div>
      </div>
    </div>
    <v-btn class="action-btn add-key-btn">
      <v-icon icon="mdi-plus" size="24" class="mr-2"></v-icon>
      Add SSH Key
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
interface SshKey {
  id: string | number
  name: string
  fingerprint: string
  addedDate: string
}
const props = defineProps<{ sshKeys: SshKey[] }>()
</script>

<style scoped>
.dashboard-card.ssh-keys-card {
  /* Minimalist card style */
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  color: var(--color-text);
  padding: 1.25rem;
  min-width: 100%;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}
.ssh-keys-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.ssh-key-item {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--rounded);
  padding: 1.2rem;
  transition: border-color var(--transition), background var(--transition);
  position: relative;
  overflow: hidden;
}
.ssh-key-item:hover {
  border-color: var(--color-accent);
}
.ssh-key-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
.ssh-key-info {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  flex: 1;
}
.ssh-key-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--rounded);
  background: none;
}
.ssh-key-name {
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.1rem;
}
.ssh-key-fingerprint, .ssh-key-date {
  font-size: 0.9rem;
  color: var(--color-text-muted);
}
.ssh-key-actions {
  display: flex;
  gap: 0.5rem;
}
.action-btn {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}
.action-btn:hover {
  background: var(--color-accent-hover);
}
.add-key-btn {
  margin-top: 0.5rem;
  font-weight: 600;
  height: 40px;
}
</style>
