<template>
  <BaseCard variant="feature" padding="large" class="base-feature-card">
    <div class="feature-icon">
      <slot name="icon">
        <v-icon color="primary" size="48">{{ icon }}</v-icon>
      </slot>
    </div>
    
    <h3 class="feature-title kubecloud-gradient">{{ title }}</h3>
    
    <p class="feature-description">
      <slot name="description">{{ description }}</slot>
    </p>
    
    <div class="feature-content">
      <slot name="content" />
    </div>
    
    <div class="feature-tags">
      <slot name="tags">
        <div 
          v-for="tag in tags" 
          :key="tag" 
          class="feature-tag"
        >
          {{ tag }}
        </div>
      </slot>
    </div>
  </BaseCard>
</template>

<script setup lang="ts">
import BaseCard from './BaseCard.vue'

interface Props {
  title: string
  description?: string
  icon?: string
  tags?: string[]
}

defineProps<Props>()
</script>

<style scoped>
.base-feature-card {
  min-height: 400px;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.feature-icon {
  margin-bottom: 1.5rem;
}

.feature-title {
  font-size: 1.8rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.feature-description {
  font-size: 1.1rem;
  color: var(--kubecloud-light-gray);
  line-height: 1.6;
  margin-bottom: 2rem;
  flex-grow: 1;
}

.feature-content {
  margin-bottom: 2rem;
}

.feature-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
}

.feature-tag {
  background: var(--kubecloud-slate);
  color: var(--kubecloud-blue);
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 500;
  border: 1px solid rgba(37, 99, 235, 0.3);
}

/* Responsive Design */
@media (max-width: 960px) {
  .base-feature-card {
    min-height: 350px;
  }
  
  .feature-title {
    font-size: 1.5rem;
  }
}

@media (max-width: 600px) {
  .base-feature-card {
    min-height: 300px;
  }
  
  .feature-title {
    font-size: 1.3rem;
  }
  
  .feature-description {
    font-size: 1rem;
  }
  
  .feature-tag {
    font-size: 0.8rem;
    padding: 0.4rem 0.8rem;
  }
}
</style> 