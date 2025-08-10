<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header row-between">
      <div>
        <h3 class="dashboard-card-title">SSH Keys</h3>
        <p class="dashboard-card-subtitle">Manage your SSH keys for secure server access</p>
      </div>
      <v-btn variant="outlined" class="add-key-btn" @click="openAddDialog">
        <v-icon icon="mdi-plus" size="20" class="mr-2"></v-icon>
        Add SSH Key
      </v-btn>
    </div>
    <div v-if="sshKeys.length === 0" class="ssh-keys-empty">
      <div class="empty-card">
        <v-icon icon="mdi-key" size="38" color="primary" class="empty-icon mb-3"></v-icon>
        <div class="empty-title">No SSH Keys Added</div>
        <div class="empty-desc">SSH keys are required to access your deployed clusters securely.<br>Add your first SSH key to get started.</div>
      </div>
    </div>
    <div v-else class="ssh-keys-list">
      <div v-for="key in sshKeys" :key="key.ID" class="ssh-key-item">
        <div class="ssh-key-content">
          <div class="ssh-key-info">
            <v-icon size="28" color="primary" class="mr-3">mdi-key</v-icon>
            <div style="width:100%">
              <div class="ssh-key-name-row">
                <span class="ssh-key-name">{{ key.name }}</span>
              </div>
              <div class="ssh-key-fingerprint-row">
                <pre class="ssh-key-fingerprint">{{ truncateKey(key.public_key) }}</pre>
                <v-tooltip location="top">
                  <template #activator="{ props }">
                    <v-btn icon small v-bind="props" @click="copyKey(key.public_key, key.ID)" title="Copy full public key">
                      <v-icon size="18">mdi-content-copy</v-icon>
                    </v-btn>
                  </template>
                  <span>Copy full public key</span>
                </v-tooltip>
                <span v-if="lastCopiedId === key.ID" class="copy-success">Copied!</span>
              </div>
              <div class="ssh-key-date">Added {{ new Date(key.created_at).toLocaleDateString() }}</div>
            </div>
          </div>
          <v-btn color="error" variant="outlined" size="small" class="action-btn" @click="handleDeleteKey(key.ID)">
            <v-icon left size="16">mdi-delete</v-icon> Remove
          </v-btn>
        </div>
      </div>
    </div>
    <v-dialog v-model="addDialog" max-width="500">
      <template #default>
        <v-card>
          <v-card-title>
            <span class="text-h6">Add SSH Key</span>
          </v-card-title>
          <v-card-text>
            <div class="mb-2 text-body-2">
              Enter a unique name and your SSH public key.<br>
              <span class="text-caption text-grey">Only valid public keys are accepted.</span>
            </div>
            <v-text-field
              ref="nameField"
              v-model="newKey.name"
              label="Key Name"
              required
              :error-messages="nameError"
              autofocus
              class="mb-4"
              @keyup.enter="handleAddKey"
              @blur="nameTouched = true"
              @input="nameTouched = true"
            />
            <v-textarea
              v-model="newKey.public_key"
              label="Public Key"
              required
              rows="4"
              :error-messages="publicKeyErrors"
              class="mb-4"
              append-inner-icon="mdi-content-paste"
              @click:append-inner="pasteFromClipboard"
              @keyup.enter="handleAddKey"
            />
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn variant="outlined" color="primary" :disabled="isAddDisabled" @click="handleAddKey">
              Add
            </v-btn>
            <v-btn variant="outlined" @click="handleCancelDialog">Cancel</v-btn>
          </v-card-actions>
        </v-card>
      </template>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { userService, type SshKey, type AddSshKeyRequest } from '../../utils/userService'

const sshKeys = ref<SshKey[]>([])
const addDialog = ref(false)
const newKey = ref<AddSshKeyRequest>({ name: '', public_key: '' })
const nameField = ref()
const lastCopiedId = ref<number | null>(null)

const isValidSshKey = (key: string): boolean =>
  /^(ssh-(rsa|ed25519|dss|ecdsa)|ecdsa-sha2-nistp\d+|sk-ecdsa-sha2-nistp\d+|sk-ssh-ed25519) [A-Za-z0-9+/=]+( [^@\s]+@[^@\s]+)?$/.test(key.trim())

const isDuplicateKey = computed(() => {
  const val = newKey.value.public_key.trim()
  return !!val && sshKeys.value.some(k => k.public_key.trim() === val)
})

const nameTouched = ref(false)
const nameError = computed(() => {
  if (!nameTouched.value) return ''
  if (!newKey.value.name) return 'Key name is required.'
  if (sshKeys.value.some(k => k.name.trim().toLowerCase() === newKey.value.name.trim().toLowerCase())) {
    return 'Key name must be unique.'
  }
  return ''
})

const publicKeyErrors = computed(() => {
  const errors: string[] = []
  if (!newKey.value.public_key) return errors
  if (!isValidSshKey(newKey.value.public_key)) errors.push('Invalid SSH public key format.')
  if (isDuplicateKey.value) errors.push('This SSH public key already exists.')
  return errors
})

const isAddDisabled = computed(() => {
  return !newKey.value.name || !newKey.value.public_key || !!nameError.value || publicKeyErrors.value.length > 0
})

async function fetchSshKeys() {
  sshKeys.value = await userService.listSshKeys()
}

async function handleAddKey() {
  if (isAddDisabled.value) return
  try {
    const added = await userService.addSshKey(newKey.value)
    sshKeys.value.push(added)
    addDialog.value = false
    newKey.value = { name: '', public_key: '' }
    nameTouched.value = false
  } catch (e: any) {
  }
}

async function handleDeleteKey(id: number) {
  await userService.deleteSshKey(id)
  sshKeys.value = sshKeys.value.filter(k => k.ID !== id)
}

async function pasteFromClipboard() {
  newKey.value.public_key = await navigator.clipboard.readText()
}


function openAddDialog() {
  addDialog.value = true
  nameTouched.value = false
  nextTick(() => {
    if (nameField.value?.focus) nameField.value.focus()
  })
}

async function copyKey(key: string, id: number) {
  await navigator.clipboard.writeText(key)
  lastCopiedId.value = id
  setTimeout(() => lastCopiedId.value = null, 1200)
}


function handleCancelDialog() {
  addDialog.value = false
  nameTouched.value = false
}

function truncateKey(key: string) {
  if (!key) return ''
  const start = key.slice(0, 25)
  const end = key.slice(-20)
  return key.length > 40 ? `${start}...${end}` : key
}

onMounted(fetchSshKeys)
</script>

<style scoped>
.row-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.ssh-keys-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-bottom: var(--space-10);
}
.ssh-key-item {
  background: rgba(30, 41, 59, 0.85);
  border: 1.5px solid #334155;
  border-radius: 12px;
  margin-bottom: 0;
  padding: 1.2rem 1.5rem;
  box-shadow: 0 2px 8px 0 rgba(0,0,0,0.07);
}
.ssh-key-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1.5rem;
}
.ssh-key-info {
  display: flex;
  align-items: flex-start;
  gap: 1.2rem;
  flex: 1;
  min-width: 0;
}
.ssh-key-name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.ssh-key-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  font-size: var(--font-size-base);
  margin: 0;
  line-height: 1.2;
}
.ssh-key-fingerprint-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.2rem;
}
.ssh-key-fingerprint {
  font-family: 'Fira Mono', 'Courier New', monospace;
  font-size: 0.97rem;
  color: #60a5fa;
  background: none;
  margin: 0.3rem 0 0.5rem 0;
  white-space: pre-wrap;
  word-break: break-all;
}
.ssh-key-date {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin: 0;
  line-height: 1.2;
}
.copy-success {
  color: #10B981;
  font-size: 0.95em;
  margin-left: 0.5em;
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
  border-color: #ef4444 !important;
  color: #ef4444 !important;
}
.add-key-btn {
  font-weight: var(--font-weight-medium);
  height: 44px;
  min-width: 180px;
}
.ssh-keys-empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 260px;
  width: 100%;
}
.empty-card {
  background: var(--color-surface-1, #18192b);
  border: 1.5px solid var(--color-border, #334155);
  border-radius: 16px;
  padding: 2.5rem 2rem 2rem 2rem;
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.10);
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 420px;
  width: 100%;
}
.empty-icon {
  margin-bottom: 0.5rem;
}
.empty-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text, #cfd2fa);
  margin-bottom: 0.5rem;
  text-align: center;
}
.empty-desc {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1.05rem;
  text-align: center;
  margin-bottom: 1.2rem;
  line-height: 1.5;
}
@media (max-width: 960px) {
  .ssh-key-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
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
}
</style>
