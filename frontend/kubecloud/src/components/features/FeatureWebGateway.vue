<template>
  <div class="feature-card card-enhanced">
    <div class="feature-icon">
      <v-icon color="primary" size="48">mdi-web</v-icon>
    </div>
    <h3 class="feature-title">Simple Web Gateway Access</h3>
    <p class="feature-description">
      Expose any service to the public web with a simple Kubernetes resource. No complex Ingress controllers. Domain and prefix-based routing is built-in.
    </p>
    <div class="feature-code">
      <pre><code>apiVersion: "gateway/v1"
kind: WebGateway
metadata:
  name: my-app-gw
spec:
  service: my-app
  domain: myapp.users.grid</code></pre>
    </div>
    <div class="feature-tags">
      <div class="feature-tag">Simple Configuration</div>
      <div class="feature-tag">Built-in Routing</div>
      <div class="feature-tag">No Ingress Controllers</div>
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
