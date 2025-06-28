<template>
  <div class="features-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title">Powerful Features</h1>
          <p class="hero-description">
            Everything you need to build, deploy, and scale your applications with confidence.
          </p>
        </div>
      </div>
    </section>
    <FeatureMyceliumPanel />
    <FeatureWebGateway />
    <FeatureMultiMaster />
    <FeatureLoadBalancing />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import FeatureMyceliumPanel from '../components/features/FeatureMyceliumPanel.vue'
import FeatureWebGateway from '../components/features/FeatureWebGateway.vue'
import FeatureMultiMaster from '../components/features/FeatureMultiMaster.vue'
import FeatureLoadBalancing from '../components/features/FeatureLoadBalancing.vue'


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
.features-view {
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

.hero-description {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  line-height: 1.7;
  opacity: 0.92;
  margin-bottom: 0;
  font-weight: 400;
}

.features-content {
  position: relative;
  z-index: 2;
  padding: 6rem 0 6rem 0;
}

.section-header {
  margin-bottom: 5rem;
}

@media (max-width: 900px) {
  .hero-section {
    padding: 3rem 0 2rem 0;
  }
  .features-content {
    padding: 3rem 0 3rem 0;
  }
  .section-header {
    margin-bottom: 2.5rem;
  }
}
@media (max-width: 600px) {
  .hero-title {
    font-size: clamp(2rem, 8vw, 3rem);
  }
  .features-content {
    padding: 2rem 0 2rem 0;
  }
}
</style>
