<template>
  <div class="feature-card card-enhanced">
    <div class="feature-icon">
      <v-icon color="primary" size="48">mdi-web</v-icon>
    </div>
    <h3 class="feature-title kubecloud-gradient">Simple Web Gateway Access</h3>
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