<template>
  <div class="feature-card card-enhanced">
    <div class="feature-icon">
      <v-icon color="primary" size="48">mdi-lan</v-icon>
    </div>
    <h3 class="feature-title kubecloud-gradient">Mycelium-Native Networking</h3>
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

<style scoped>
.feature-code {
  background: var(--kubecloud-slate);
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 2rem;
  border: 1px solid rgba(37, 99, 235, 0.2);
}

.feature-code pre {
  margin: 0;
  color: var(--kubecloud-light-gray);
  font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
  font-size: 0.9rem;
  line-height: 1.4;
}

.feature-code code {
  color: var(--kubecloud-blue);
}
</style> 