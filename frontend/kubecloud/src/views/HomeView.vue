<script setup lang="ts">
// Add scroll animation observer
import { onMounted } from 'vue'
import FeatureGlobe from '@/components/features/FeatureGlobe.vue'

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

const features = [
  {
    icon: 'mdi-cloud-outline',
    title: 'Cloud-Native Architecture',
    description: 'Built for the cloud with support for all major cloud providers and on-premise deployments.'
  },
  {
    icon: 'mdi-shield-check-outline',
    title: 'Enterprise Security',
    description: 'Advanced security features including RBAC, network policies, and compliance monitoring.'
  },
  {
    icon: 'mdi-chart-line',
    title: 'Real-time Monitoring',
    description: 'Comprehensive monitoring and alerting with detailed metrics and performance insights.'
  },
  {
    icon: 'mdi-rocket-launch-outline',
    title: 'One-Click Deployments',
    description: 'Streamlined deployment process with automated CI/CD pipelines and rollback capabilities.'
  },
  {
    icon: 'mdi-account-group-outline',
    title: 'Team Collaboration',
    description: 'Built-in collaboration tools for teams with role-based access and shared workspaces.'
  },
  {
    icon: 'mdi-cog-outline',
    title: 'Advanced Configuration',
    description: 'Flexible configuration management with support for Helm charts and custom resources.'
  }
]
</script>

<template>
  <div class="home-view">
    <!-- Hero Globe Section -->
    <section class="hero-globe-section">
      <div class="hero-globe-content">
        <div class="hero-globe-text">
          <h1 class="hero-title">KubeCloud</h1>
          <p class="hero-subtitle">The next generation Kubernetes platform for teams and enterprises</p>
        </div>
        <FeatureGlobe />
      </div>
    </section>
    <!-- Features Section -->
    <section class="home-section section-padding">
      <div class="container-padding">
        <div class="section-header text-center mb-8">
          <h2 class="section-title">
            Everything You Need to Succeed
          </h2>
          <p class="section-subtitle">
            Powerful tools and features designed for modern cloud-native applications
          </p>
        </div>

        <v-row>
          <v-col cols="12" md="4" v-for="feature in features" :key="feature.title">
            <div class="home-card card-enhanced">
              <div class="home-icon">
                <v-icon :icon="feature.icon" size="48" color="primary"></v-icon>
              </div>
              <h3 class="home-title">{{ feature.title }}</h3>
              <p class="home-description">{{ feature.description }}</p>
            </div>
          </v-col>
        </v-row>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="cta-section section-padding">
      <div class="container-padding">
        <div class="cta-content text-center">
          <h2 class="cta-title">
            Ready to Transform Your Kubernetes Experience?
          </h2>
          <p class="cta-description">
            Join thousands of developers and DevOps engineers who trust KubeCloud for their production workloads.
          </p>
          <v-btn
            color="secondary"
            size="x-large"
            variant="elevated"
            to="/sign-up"
            class="cta-btn btn-enhanced kubecloud-glow-orange"
          >
            Start Your Free Trial
          </v-btn>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home-view {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
}

/* Features Section */
.home-section {
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

.home-card {
  padding: 2rem;
  text-align: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.home-card:hover {
  transform: translateY(-8px);
}

.home-icon {
  margin-bottom: 1.5rem;
}

.home-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 1rem;
}

.home-description {
  color: #CBD5E1;
  line-height: 1.6;
  opacity: 0.8;
}

/* CTA Section */
.cta-section {
  position: relative;
  z-index: 2;
  padding: 4rem 0;
}

.cta-content {
  max-width: 600px;
  margin: 0 auto;
}

.cta-title {
  font-size: clamp(1.75rem, 3.5vw, 2.5rem);
  font-weight: 700;
  margin-bottom: 1rem;
  line-height: 1.3;
  color: #FFFFFF;
}

.cta-description {
  font-size: 1.125rem;
  color: #CBD5E1;
  margin-bottom: 2rem;
  opacity: 0.8;
  line-height: 1.6;
}

.cta-btn {
  font-size: 1.125rem;
  padding: 1rem 2rem;
}

/* Responsive Design */
@media (max-width: 768px) {
  .home-card {
    padding: 1.5rem;
  }
}

@media (max-width: 480px) {
  .home-section,
  .cta-section {
    padding: 2rem 0;
  }
}

.hero-globe-section {
  width: 100%;
  min-height: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, #0a192f 60%, #1e293b 100%);
  position: relative;
  z-index: 1;
  padding-top: 2rem;
  padding-bottom: 2rem;
}
.hero-globe-content {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 3rem;
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
}
.hero-globe-text {
  flex: 1;
  min-width: 220px;
  text-align: left;
}
.hero-title {
  font-size: clamp(2.5rem, 6vw, 4rem);
  font-weight: 800;
  color: #fff;
  margin-bottom: 1rem;
  letter-spacing: -1px;
}
.hero-subtitle {
  font-size: 1.25rem;
  color: #60a5fa;
  opacity: 0.9;
  max-width: 400px;
}
@media (max-width: 900px) {
  .hero-globe-content {
    flex-direction: column;
    gap: 2rem;
    text-align: center;
  }
  .hero-globe-text {
    text-align: center;
  }
}
</style>
