<template>
  <div class="features-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title kubecloud-gradient kubecloud-glow-blue">
            Powerful Features
          </h1>
          <p class="hero-description">
            Everything you need to build, deploy, and scale your applications with confidence.
          </p>
        </div>
      </div>
    </section>

    <!-- Features Content -->
    <section class="features-content section-padding">
      <div class="container-padding">
        <div class="section-header text-center mb-8">
          <h2 class="section-title kubecloud-gradient kubecloud-glow-blue">
            Enterprise-Grade Capabilities
          </h2>
          <p class="section-subtitle">
            Advanced features designed for modern cloud-native applications
          </p>
        </div>
        
        <v-row>
          <v-col cols="12" md="4" v-for="feature in features" :key="feature.title">
            <component :is="feature.component" class="feature-card card-enhanced fade-in" />
          </v-col>
        </v-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import FeatureMyceliumNetworking from '../components/features/FeatureMyceliumNetworking.vue'
import FeatureWebGateway from '../components/features/FeatureWebGateway.vue'
import FeatureGpuOnDemand from '../components/features/FeatureGpuOnDemand.vue'

const features = [
  { title: 'Mycelium Networking', component: FeatureMyceliumNetworking },
  { title: 'Web Gateway', component: FeatureWebGateway },
  { title: 'GPU On-Demand', component: FeatureGpuOnDemand }
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

function logout() {
  // Logout logic here
}
</script>

<style scoped>
.features-view {
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

/* Features Content */
.features-content {
  position: relative;
  z-index: 2;
  padding: 4rem 0;
}

.section-header {
  margin-bottom: 3rem;
}

.section-title {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 700;
  margin-bottom: 1rem;
  line-height: 1.3;
}

.section-subtitle {
  font-size: 1.125rem;
  color: #CBD5E1;
  opacity: 0.8;
  max-width: 600px;
  margin: 0 auto;
}

.feature-card {
  height: 100%;
  padding: 2rem;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.feature-card:hover {
  transform: translateY(-8px);
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
  
  .features-content {
    padding: 2rem 0;
  }
  
  .feature-card {
    padding: 1.5rem;
  }
}

@media (max-width: 480px) {
  .hero-section {
    min-height: 40vh;
  }
  
  .features-content {
    padding: 1.5rem 0;
  }
}
</style>
