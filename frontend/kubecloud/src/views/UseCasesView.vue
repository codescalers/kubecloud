<template>
  <div class="use-cases-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title">Use Cases</h1>
          <p class="section-subtitle">Explore how KubeCloud empowers different industries and teams.</p>
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
              <h2 class="section-title">{{ useCase.title }}</h2>
              <p class="use-case-description card-subtitle">{{ useCase.description }}</p>
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
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
}

.hero-section {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
  padding: 6rem 0 4rem 0;
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
  font-size: clamp(2.5rem, 5vw, 3.5rem);
  font-weight: 500;
  margin-bottom: 2.5rem;
  line-height: 1.1;
  letter-spacing: -1px;
  color: #fff;
}

.section-subtitle {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  line-height: 1.7;
  opacity: 0.92;
  margin-bottom: 0;
  font-weight: 400;
}

.use-cases-content {
  position: relative;
  z-index: 2;
  padding: 6rem 0 6rem 0;
}


.use-case-card {
  height: 100%;
  padding: 2.5rem;
  text-align: center;
  background: rgba(10, 25, 47, 0.85);
  border-radius: 1.5rem;
  box-shadow: 0 4px 20px rgba(96, 165, 250, 0.08);
  color: #CBD5E1;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.use-case-card:hover {
  transform: translateY(-10px) scale(1.03);
  box-shadow: 0 8px 32px rgba(96, 165, 250, 0.18);
}
.use-case-icon {
  margin-bottom: 2rem;
}

.use-case-features {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
}

.feature-tag {
  background: transparent;
  color: #fff;
  padding: 0.4rem 0.8rem;
  border-radius: 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.feature-tag:hover {
  border-color: rgba(255, 255, 255, 0.6);
  background: rgba(255, 255, 255, 0.05);
}

@media (max-width: 900px) {
  .hero-section {
    padding: 3rem 0 2rem 0;
  }
  .use-cases-content {
    padding: 3rem 0 3rem 0;
  }
  .use-case-card {
    padding: 1.5rem;
  }
}
@media (max-width: 600px) {
  .hero-title {
    font-size: clamp(2rem, 8vw, 3rem);
  }
  .use-cases-content {
    padding: 2rem 0 2rem 0;
  }
}
</style>
