<template>
  <section class="d-flex align-center justify-center use-cases-section">
    <v-container fluid class="py-16">
      <!-- Hero Section -->
      <v-container fluid class="pt-16 pt-md-20">
        <v-row justify="center" align="center" class="text-center">
          <v-col cols="12" sm="10" md="8" lg="6">
            <h1 class="text-h4 text-md-h2 font-weight-medium mb-8 text-white">
              Use Cases
            </h1>
            <p class="text-h6 text-blue-lighten-2 line-height-1-7 opacity-92">
              Explore how Mycelium Cloud empowers different industries and teams.
            </p>
          </v-col>
        </v-row>
      </v-container>

      <!-- Use Cases Content -->
      <v-container class="py-12 py-md-16">
        <v-row justify="center" class="mx-auto" style="max-width: 1200px;">
          <v-col 
            cols="12" 
            sm="6" 
            md="4" 
            v-for="useCase in useCases" 
            :key="useCase.title"
            class="mb-6"
          >
            <v-card 
              class="use-case-card fade-in h-100 d-flex flex-column"
              elevation="8"
              rounded="xl"
            >
              <v-card-text class="text-center pa-8 pa-md-10 d-flex flex-column flex-grow-1">
                <div class="mb-6">
                  <v-icon :icon="useCase.icon" size="48" color="primary"></v-icon>
                </div>
                
                <h2 class="text-h4 text-md-h5 font-weight-medium mb-4 text-white">
                  {{ useCase.title }}
                </h2>
                
                <p class="text-body-1 text-subtitle-1 font-weight-normal text-grey-300 mb-6 flex-grow-1">
                  {{ useCase.description }}
                </p>
                
                <div class="d-flex flex-wrap justify-center gap-2 mt-auto">
                  <v-chip
                    v-for="feature in useCase.features"
                    :key="feature"
                    size="small"
                    class="ma-1"
                  >
                    {{ feature }}
                  </v-chip>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-container>
  </section>
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

// Simple scroll animation observer for fade-in effect
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
/* Background gradient for the entire section */
.use-cases-section {
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), 
              radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
}

/* Use case card styling */
.use-case-card {
  background: rgba(255, 255, 255, 0.05) !important;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s ease;
}

.use-case-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3) !important;
}


/* Fade-in animation */
.fade-in {
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.6s ease;
}

.fade-in.visible {
  opacity: 1;
  transform: translateY(0);
}

/* Custom gap utility for chips */
.gap-2 {
  gap: 8px;
}
</style>

