<template>
  <div class="feature-card card-enhanced">
    <div class="feature-icon">
      <v-icon color="primary" size="48">mdi-lan</v-icon>
    </div>
    <h3 class="feature-title">Mycelium-Native Networking</h3>
    <p class="feature-description">
      Every pod gets a private, end-to-end encrypted Mycelium IP address. Use standard Kubernetes Network Policies to create whitelists and secure your workloads.
    </p>
    <div class="feature-code">
      <pre><code>apiVersion: "v1"
kind: NetworkPolicy
metadata:
  name: allow-ingress
spec:
  podSelector: ...</code></pre>
    </div>
    <div class="feature-tags">
      <div class="feature-tag">End-to-End Encryption</div>
      <div class="feature-tag">Network Policies</div>
      <div class="feature-tag">Private IPs</div>
    </div>
  </div>
</template>

<script setup lang="ts">
// No logic needed
</script>

<script lang="ts">
export default {}
</script>

<style scoped>
.feature-code {
  background: var(--kubecloud-surface-dark);
  border-radius: 10px;
  padding: 1rem;
  margin-bottom: 2rem;
  border: 1px solid var(--kubecloud-border-dark);
}

.feature-code pre {
  margin: 0;
  color: var(--kubecloud-light-gray);
  font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
  font-size: 0.95rem;
  line-height: 1.5;
}

.feature-code code {
  color: var(--kubecloud-primary);
}

.feature-tag {
  display: inline-block;
  background: rgba(59, 130, 246, 0.15);
  color: #FFFFFF;
  border-radius: 6px;
  padding: 0.25rem 0.75rem;
  font-size: 0.85rem;
  margin-right: 0.5rem;
  margin-bottom: 0.5rem;
  border: 1px solid rgba(59, 130, 246, 0.3);
  font-weight: 500;
  letter-spacing: 0.01em;
  transition: all 0.2s ease;
}

.feature-tag:hover {
  background: rgba(59, 130, 246, 0.25);
  border-color: rgba(59, 130, 246, 0.5);
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.feature-tags {
  margin-top: 1rem;
}
</style>
