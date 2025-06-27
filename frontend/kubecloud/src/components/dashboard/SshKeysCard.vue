<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">SSH Keys</h3>
          <p class="dashboard-card-subtitle">Manage your SSH keys for secure server access</p>
        </div>
      </div>
    </div>
    <div class="ssh-keys-list">
      <div v-for="key in sshKeys" :key="key.id" class="ssh-key-item list-item-interactive">
        <div class="ssh-key-content">
          <div class="ssh-key-info">
            <div class="ssh-key-icon">
              <v-icon size="24" color="primary">mdi-key</v-icon>
            </div>
            <div class="ssh-key-details">
              <h3 class="ssh-key-name">{{ key.name }}</h3>
              <p class="ssh-key-fingerprint">{{ key.fingerprint }}</p>
              <p class="ssh-key-date">Added {{ key.addedDate }}</p>
            </div>
          </div>
          <div class="ssh-key-actions">
            <v-btn variant="outlined" size="small" class="action-btn">Edit</v-btn>
            <v-btn variant="outlined" size="small" class="action-btn">Remove</v-btn>
          </div>
        </div>
      </div>
    </div>
    <div class="add-key-section">
      <v-btn variant="outlined" class="add-key-btn action-btn">
        <v-icon icon="mdi-plus" size="20" class="mr-2"></v-icon>
        Add SSH Key
      </v-btn>
    </div>
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
.ssh-keys-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  margin-bottom: var(--space-10);
}

.ssh-key-item {
  padding: var(--space-8);
  border-radius: var(--radius-lg);
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid var(--color-border);
  transition: background 0.18s, border-color 0.18s;
}

.ssh-key-item:hover {
  background: rgba(30, 41, 59, 0.85);
  border-color: var(--color-border-light);
}

.ssh-key-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-12);
}

.ssh-key-info {
  display: flex;
  align-items: center;
  gap: var(--space-6);
  flex: 1;
  min-width: 0;
}

.ssh-key-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-md);
  background: rgba(59, 130, 246, 0.07);
  color: var(--color-primary);
  flex-shrink: 0;
}

.ssh-key-details {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  min-width: 0;
}

.ssh-key-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  font-size: var(--font-size-base);
  margin: 0;
  line-height: 1.2;
}

.ssh-key-fingerprint {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
  opacity: 0.9;
  margin: 0;
  line-height: 1.3;
  font-family: 'Courier New', monospace;
  word-break: break-all;
}

.ssh-key-date {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin: 0;
  line-height: 1.2;
}

.ssh-key-actions {
  display: flex;
  gap: var(--space-4);
  flex-shrink: 0;
}

.action-btn {
  background: transparent !important;
  border: 1px solid var(--color-border) !important;
  color: var(--color-text) !important;
  font-weight: var(--font-weight-medium);
  transition: background 0.18s, border-color 0.18s;
  white-space: nowrap;
  box-shadow: none !important;
}

.action-btn:hover {
  background: rgba(59, 130, 246, 0.07) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}

.add-key-section {
  display: flex;
  justify-content: center;
  padding-top: var(--space-8);
  border-top: 1px solid var(--color-border);
}

.add-key-btn {
  font-weight: var(--font-weight-medium);
  height: 44px;
  min-width: 220px;
}

@media (max-width: 960px) {
  .ssh-key-content {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-6);
  }

  .ssh-key-actions {
    align-self: stretch;
    justify-content: flex-end;
  }
}

@media (max-width: 600px) {
  .ssh-key-item {
    padding: var(--space-5);
  }

  .ssh-key-info {
    gap: var(--space-4);
  }

  .ssh-key-icon {
    width: 44px;
    height: 44px;
  }

  .ssh-key-details {
    gap: var(--space-2);
  }

  .ssh-key-actions {
    gap: var(--space-2);
  }
  
  .add-key-btn {
    min-width: 100%;
  }
}
</style>
