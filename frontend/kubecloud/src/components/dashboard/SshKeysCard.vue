<template>
  <v-card class="pa-8 profile-card ssh-keys-card card-enhanced">
    <h2 class="display-3 mb-6 neon-glow-cyan">SSH Keys</h2>
    <div class="ssh-keys-list mb-6">
      <div v-for="key in sshKeys" :key="key.id" class="ssh-key-item">
        <div class="ssh-key-content">
          <div class="ssh-key-info">
            <div class="ssh-key-icon">
              <v-icon size="32" color="primary">mdi-key</v-icon>
            </div>
            <div class="ssh-key-details">
              <h3 class="ssh-key-name">{{ key.name }}</h3>
              <p class="ssh-key-fingerprint">{{ key.fingerprint }}</p>
              <p class="ssh-key-date">Added {{ key.addedDate }}</p>
            </div>
          </div>
          <div class="ssh-key-actions">
            <v-btn variant="text" color="primary" size="large" class="action-btn neon-hover-violet">Edit</v-btn>
            <v-btn variant="text" color="error" size="large" class="action-btn neon-hover-rose">Remove</v-btn>
          </div>
        </div>
      </div>
    </div>
    <v-btn variant="elevated" color="primary" size="large" class="add-key-btn neon-glow-cyan">
      <v-icon icon="mdi-plus" size="24" class="mr-2"></v-icon>
      Add SSH Key
    </v-btn>
  </v-card>
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
.ssh-keys-card {
  border-radius: 24px;
  min-width: 100%;
}

.ssh-keys-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.ssh-key-item {
  background: rgba(8, 145, 178, 0.05);
  border: 1px solid rgba(8, 145, 178, 0.2);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.ssh-key-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--kubecloud-cyan), var(--kubecloud-cyan-light));
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.ssh-key-item:hover {
  background: rgba(8, 145, 178, 0.1);
  border: 1px solid var(--kubecloud-cyan);
  box-shadow: var(--shadow-kubecloud-cyan);
  transform: translateY(-1px);
}

.ssh-key-item:hover::before {
  opacity: 1;
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
  gap: 1rem;
  flex: 1;
}

.ssh-key-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: rgba(8, 145, 178, 0.1);
  border: 1px solid rgba(8, 145, 178, 0.3);
  transition: all var(--transition-normal);
}

.ssh-key-item:hover .ssh-key-icon {
  background: rgba(8, 145, 178, 0.2);
  border: 1px solid var(--kubecloud-cyan);
  box-shadow: var(--shadow-kubecloud-cyan);
  transform: scale(1.02);
}

.ssh-key-details {
  flex: 1;
}

.ssh-key-name {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.25rem;
  font-weight: 600;
  color: #fff;
  margin: 0 0 0.25rem 0;
  letter-spacing: -0.01em;
}

.ssh-key-fingerprint {
  font-family: 'Fira Mono', 'Menlo', 'Monaco', 'Consolas', monospace;
  font-size: 0.875rem;
  color: var(--kubecloud-cyan);
  margin: 0 0 0.25rem 0;
  font-weight: 500;
  letter-spacing: 0.02em;
  text-shadow: 0 0 4px rgba(8, 145, 178, 0.3);
}

.ssh-key-date {
  font-size: 0.875rem;
  color: #E0E7EF;
  margin: 0;
  font-weight: 400;
  opacity: 0.8;
}

.ssh-key-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.action-btn {
  font-weight: 600;
  font-size: 1rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: transparent;
  border: 1px solid transparent;
  transition: all var(--transition-normal);
  color: #fff !important;
}

.action-btn:hover {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

.add-key-btn {
  font-weight: 600;
  font-size: 1.125rem;
  padding: 1rem 1.5rem;
  border-radius: 12px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: linear-gradient(135deg, var(--kubecloud-cyan), var(--kubecloud-cyan-light)) !important;
  border: 1px solid var(--kubecloud-cyan);
  box-shadow: var(--shadow-kubecloud-cyan);
  color: var(--kubecloud-white) !important;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
}

.add-key-btn:hover {
  background: linear-gradient(135deg, var(--kubecloud-cyan-light), var(--kubecloud-cyan)) !important;
  box-shadow: var(--shadow-hover-cyan);
  transform: translateY(-2px);
}

@media (max-width: 960px) {
  .ssh-keys-card {
    padding: 2rem !important;
  }
  
  .ssh-key-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .ssh-key-actions {
    align-self: flex-end;
  }
  
  .add-key-btn {
    font-size: 1rem;
    padding: 0.75rem 1.25rem;
    min-height: 44px;
  }
}

@media (max-width: 600px) {
  .ssh-keys-card {
    padding: 1.5rem !important;
  }
  
  .ssh-key-item {
    padding: 1rem;
  }
  
  .ssh-key-icon {
    width: 48px;
    height: 48px;
  }
  
  .ssh-key-name {
    font-size: 1.125rem;
  }
  
  .ssh-key-fingerprint {
    font-size: 0.8rem;
  }
  
  .action-btn {
    font-size: 0.9rem;
    padding: 0.375rem 0.75rem;
  }
  
  .add-key-btn {
    font-size: 0.9rem;
    padding: 0.5rem 1rem;
    min-height: 40px;
  }
}
</style>
