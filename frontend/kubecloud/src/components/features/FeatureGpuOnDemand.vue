<template>
  <div class="feature-card card-enhanced">
    <div class="feature-icon">
      <v-icon color="primary" size="48">mdi-chip</v-icon>
    </div>
    <h3 class="feature-title kubecloud-gradient">GPU on Demand</h3>
    <p class="feature-description">
      Get instant access to powerful computing for your projects. Run demanding apps easily and scale up when you need more power. Pay only for what you use.
    </p>
    <div class="feature-diagram">
      <div class="diagram-illustration">
        <div class="diagram-flow">
          <span class="diagram-item">GPU</span>
          <span class="diagram-arrow">⟶</span>
          <span class="diagram-item">K8s Node</span>
          <span class="diagram-arrow">⟶</span>
          <span class="diagram-item">Pod</span>
        </div>
        <div class="diagram-labels">
          <span class="diagram-label">Hardware</span>
          <span class="diagram-label">Compute</span>
          <span class="diagram-label">Your App</span>
        </div>
      </div>
    </div>
    <div class="feature-tags">
      <div class="feature-tag">Instant Access</div>
      <div class="feature-tag">Full Passthrough</div>
      <div class="feature-tag">Pay-per-Use</div>
    </div>
  </div>
</template>

<script setup lang="ts">
// No logic needed
</script>

<style scoped>
.feature-diagram {
  background: var(--kubecloud-slate);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  border: 1px solid rgba(37, 99, 235, 0.2);
}

.diagram-illustration {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.diagram-flow {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 1.1rem;
  margin-bottom: 1rem;
}

.diagram-item {
  background: var(--kubecloud-blue);
  color: var(--kubecloud-white);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.9rem;
}

.diagram-arrow {
  color: var(--kubecloud-orange);
  font-size: 1.3rem;
  font-weight: bold;
}

.diagram-labels {
  display: flex;
  justify-content: space-between;
  width: 100%;
  max-width: 320px;
  font-size: 0.85rem;
  color: var(--kubecloud-light-gray);
}

.diagram-label {
  flex: 1;
  text-align: center;
}

/* Responsive Design */
@media (max-width: 960px) {
  .diagram-flow {
    gap: 0.7rem;
    font-size: 1rem;
  }
  
  .diagram-item {
    font-size: 0.8rem;
    padding: 0.4rem 0.8rem;
  }
}

@media (max-width: 600px) {
  .diagram-flow {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .diagram-arrow {
    transform: rotate(90deg);
  }
  
  .diagram-labels {
    flex-direction: column;
    gap: 0.5rem;
  }
}
</style>
