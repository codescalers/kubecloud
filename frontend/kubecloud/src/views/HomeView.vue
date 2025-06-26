<script setup lang="ts">
// Add scroll animation observer
import { onMounted } from 'vue'

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
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text">
          <h1 class="hero-title kubecloud-gradient kubecloud-glow-blue">
            Build, Deploy, Scale
          </h1>
          <h2 class="hero-subtitle">
            That's how real engineers manage Kubernetes clusters
          </h2>
          <p class="hero-description">
            Experience the world's most advanced Kubernetes management platform. Deploy, monitor, and scale your applications with enterprise-grade reliability and developer-friendly simplicity.
          </p>
          
          <div class="hero-cta">
            <v-btn 
              color="secondary" 
              size="x-large" 
              variant="elevated" 
              to="/sign-up" 
              class="cta-primary btn-enhanced kubecloud-glow-orange"
            >
              Get Started Free
            </v-btn>
            <v-btn 
              color="primary" 
              size="x-large" 
              variant="outlined" 
              to="/features" 
              class="cta-secondary btn-enhanced kubecloud-hover-blue"
            >
              Explore Features
            </v-btn>
          </div>
          
          <div class="hero-stats">
            <div class="stat-item">
              <div class="stat-number kubecloud-gradient">10K+</div>
              <div class="stat-label">Active Clusters</div>
            </div>
            <div class="stat-item">
              <div class="stat-number kubecloud-gradient">99.9%</div>
              <div class="stat-label">Uptime</div>
            </div>
            <div class="stat-item">
              <div class="stat-number kubecloud-gradient">24/7</div>
              <div class="stat-label">Support</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="home-section section-padding">
      <div class="container-padding">
        <div class="section-header text-center mb-8">
          <h2 class="section-title kubecloud-gradient kubecloud-glow-blue">
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
          <h2 class="cta-title kubecloud-gradient kubecloud-glow-blue">
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

/* Hero Section */
.hero-section {
  min-height: 100vh;
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
  margin-bottom: 1rem;
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.hero-subtitle {
  font-size: clamp(1.25rem, 2.5vw, 1.75rem);
  color: #E0E7EF;
  font-weight: 500;
  margin-bottom: 1.5rem;
  opacity: 0.9;
}

.hero-description {
  font-size: 1.125rem;
  color: #CBD5E1;
  line-height: 1.7;
  margin-bottom: 2.5rem;
  opacity: 0.8;
}

.hero-cta {
  display: flex;
  gap: 1rem;
  justify-content: center;
  margin-bottom: 3rem;
  flex-wrap: wrap;
}

.hero-stats {
  display: flex;
  justify-content: center;
  gap: 3rem;
  flex-wrap: wrap;
}

.stat-item {
  text-align: center;
}

.stat-number {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.875rem;
  color: #CBD5E1;
  opacity: 0.7;
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
  .hero-cta {
    flex-direction: column;
    align-items: center;
  }
  
  .hero-stats {
    gap: 2rem;
  }
  
  .home-card {
    padding: 1.5rem;
  }
}

@media (max-width: 480px) {
  .hero-section {
    padding: 1rem 0;
  }
  
  .hero-stats {
    gap: 1.5rem;
  }
  
  .stat-number {
    font-size: 1.5rem;
  }
  
  .home-section,
  .cta-section {
    padding: 2rem 0;
  }
}
</style>
