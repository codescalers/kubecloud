<template>
  <div class="feature-card card-enhanced">
    <div class="feature-icon">
      <v-icon color="primary" size="48">mdi-chip</v-icon>
    </div>
    <!-- Removed gaming-like class -->
    <h3 class="feature-title">GPU on Demand</h3>
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

<script lang="ts">
export default {}
</script>

<style scoped>
.feature-diagram {
  background: var(--kubecloud-surface-light); /* Use lighter surface */
  border-radius: var(--rounded-md); /* Use standardized border radius */
  padding: 1.2rem; /* Adjusted padding */
  margin-bottom: 1.5rem; /* Adjusted margin */
  border: 1px solid var(--kubecloud-border); /* Use border color */
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
  gap: 0.8rem; /* Adjusted gap */
  font-size: var(--text-base); /* Use standardized font size */
  margin-bottom: 0.8rem; /* Adjusted margin */
}

.diagram-item {
  background: var(--kubecloud-primary); /* Use primary color */
  color: var(--kubecloud-text-primary); /* Use primary text color */
  padding: 0.4rem 0.8rem; /* Adjusted padding */
  border-radius: var(--rounded-sm); /* Use standardized border radius */
  font-weight: var(--font-medium); /* Use standardized font weight */
  font-size: var(--text-sm); /* Use standardized font size */
}

.diagram-arrow {
  color: var(--kubecloud-text-secondary); /* Use secondary text color */
  font-size: var(--text-lg); /* Adjusted font size */
  font-weight: var(--font-bold); /* Use standardized font weight */
}

.diagram-labels {
  display: flex;
  justify-content: space-between;
  width: 100%;
  max-width: 300px; /* Adjusted max-width */
  font-size: var(--text-xs); /* Use standardized font size */
  color: var(--kubecloud-text-secondary); /* Use secondary text color */
}

.diagram-label {
  flex: 1;
  text-align: center;
  color: var(--kubecloud-text-secondary);
}

/* Responsive Design - Align with updated sizes and padding */
@media (max-width: 960px) {
  .feature-diagram {
    padding: 1rem; /* Adjusted padding */
    margin-bottom: 1.2rem; /* Adjusted margin */
  }
  .diagram-flow {
    gap: 0.6rem; /* Adjusted gap */
    font-size: var(--text-sm); /* Adjusted font size */
    margin-bottom: 0.6rem; /* Adjusted margin */
  }
  .diagram-item {
    font-size: var(--text-xs); /* Adjusted font size */
    padding: 0.3rem 0.6rem; /* Adjusted padding */
  }
  .diagram-arrow {
    font-size: var(--text-base); /* Adjusted font size */
  }
  .diagram-labels {
    max-width: 250px; /* Adjusted max-width */
    font-size: var(--text-xs); /* Adjusted font size */
  }
}

@media (max-width: 600px) {
  .feature-diagram {
    padding: 0.8rem; /* Adjusted padding */
    margin-bottom: 1rem; /* Adjusted margin */
  }
  .diagram-flow {
    flex-direction: column;
    gap: 0.4rem; /* Adjusted gap */
    font-size: var(--text-sm); /* Adjusted font size */
  }
  .diagram-arrow {
    transform: rotate(90deg);
    font-size: var(--text-base); /* Adjusted font size */
  }
  .diagram-labels {
    flex-direction: column;
    gap: 0.4rem; /* Adjusted gap */
    max-width: 100%;
  }
}

/* Feature Tags */
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
