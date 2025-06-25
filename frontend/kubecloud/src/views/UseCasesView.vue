<template>
  <div class="use-cases-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title kubecloud-gradient kubecloud-glow-blue">
            Use Cases
          </h1>
          <p class="hero-description">
            Discover how KubeCloud powers real-world applications across industries.
          </p>
        </div>
      </div>
    </section>

    <!-- Use Cases Content -->
    <section class="use-cases-content section-padding">
      <div class="container-padding">
        <v-row>
          <v-col cols="12" md="6" lg="4" v-for="useCase in useCases" :key="useCase.title">
            <div class="use-case-card card-enhanced fade-in">
              <div class="use-case-icon">
                <v-icon :icon="useCase.icon" size="48" color="primary"></v-icon>
              </div>
              <h3 class="use-case-title">{{ useCase.title }}</h3>
              <p class="use-case-description">{{ useCase.description }}</p>
              <div class="use-case-features">
                <div class="feature-tag" v-for="feature in useCase.features" :key="feature">
                  {{ feature }}
                </div>
              </div>
            </div>
          </v-col>
        </v-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'

const useCases = [
  {
    icon: 'mdi-brain',
    title: 'AI/ML Training',
    description: 'Run GPU-accelerated workloads for deep learning and data science on demand.',
    features: ['GPU Acceleration', 'Scalable Compute', 'Cost Optimization']
  },
  {
    icon: 'mdi-domain',
    title: 'Enterprise Kubernetes',
    description: 'Deploy production-grade clusters with full control and no vendor lock-in.',
    features: ['High Availability', 'Security', 'Compliance']
  },
  {
    icon: 'mdi-router-wireless',
    title: 'Edge & IoT',
    description: 'Leverage global nodes for low-latency edge computing and IoT applications.',
    features: ['Low Latency', 'Global Distribution', 'Real-time Processing']
  }
]

// Add scroll animation observer
onMounted(() => {
  const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px -50px 0px'
  }

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, observerOptions)

  document.querySelectorAll('.fade-in').forEach(el => {
    observer.observe(el)
  })
})
</script>

<style scoped>
.use-cases-view {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
}

/* Hero Section */
.hero-section {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
  padding: 2rem 0;
}

.hero-content {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}

.hero-text {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: clamp(2.5rem, 5vw, 4rem);
  font-weight: 700;
  margin-bottom: 1.5rem;
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.hero-description {
  font-size: 1.25rem;
  color: #CBD5E1;
  line-height: 1.7;
  opacity: 0.8;
}

/* Use Cases Content */
.use-cases-content {
  position: relative;
  z-index: 2;
  padding: 4rem 0;
}

.use-case-card {
  height: 100%;
  padding: 2.5rem;
  text-align: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.use-case-card:hover {
  transform: translateY(-8px);
}

.use-case-icon {
  margin-bottom: 1.5rem;
}

.use-case-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 1rem;
}

.use-case-description {
  color: #CBD5E1;
  line-height: 1.6;
  margin-bottom: 2rem;
  opacity: 0.8;
}

.use-case-features {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
}

.feature-tag {
  background: rgba(79, 70, 229, 0.1);
  color: #7C3AED;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid rgba(79, 70, 229, 0.3);
}

/* Responsive Design */
@media (max-width: 768px) {
  .hero-section {
    min-height: 50vh;
    padding: 1rem 0;
  }
  
  .hero-description {
    font-size: 1.125rem;
  }
  
  .use-cases-content {
    padding: 2rem 0;
  }
  
  .use-case-card {
    padding: 2rem;
  }
}

@media (max-width: 480px) {
  .hero-section {
    min-height: 40vh;
  }
  
  .use-cases-content {
    padding: 1.5rem 0;
  }
  
  .use-case-card {
    padding: 1.5rem;
  }
  
  .feature-tag {
    font-size: 0.8rem;
    padding: 0.4rem 0.8rem;
  }
}
</style>
