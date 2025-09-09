<script setup lang="ts">
import { onMounted, ref, computed, defineAsyncComponent } from 'vue'
const FeatureGlobe = defineAsyncComponent(() => import('../components/features/FeatureGlobe.vue'))
import { useUserStore } from '../stores/user'
import { statsService } from '../utils/statsService'
import { useGridNodes } from '../composables/useGridNodes'
import { formatStatsForCards, type FormattedStat } from '../utils/formatUtils'
import { processGridNodesForGlobe } from '../utils/globeUtils'

const userStore = useUserStore()
const globeSize = ref(900)
const isLoading = ref(true)
const { gridNodes, fetchGridNodes } = useGridNodes()

const displayStats = ref<FormattedStat[]>([])
const isGridStatsLoading = ref(false)
const gridCards = computed(() => displayStats.value)
const processedGlobeData = computed(() => {
  try {
    const result = processGridNodesForGlobe(gridNodes.value)
    if (result.nodes.length > 0) {
      return result
    }
    return { nodes: [], labels: [] }
  } catch (error) {
    console.error('Error processing globe nodes:', error)
    return { nodes: [], labels: [] }
  }
})

const globeNodes = computed(() => {
  const data = processedGlobeData.value
  return data.nodes.length > 0 ? data.nodes : undefined
})

const globeLabels = computed(() => {
  const data = processedGlobeData.value
  return data.labels.length > 0 ? data.labels : undefined
})

async function fetchStats() {
  try {
    const stats = await statsService.getStats()
    displayStats.value = formatStatsForCards(stats)
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  } finally {
    isLoading.value = false
  }
}

function updateGlobeSize() {
  globeSize.value = Math.max(600, Math.min(800, Math.floor(window.innerWidth * 0.6)))
}

onMounted(async () => {
  updateGlobeSize()
  window.addEventListener('resize', updateGlobeSize)
  await fetchStats()
  try {
    await fetchGridNodes({ healthy: true, size: 1000 })
  } catch (error) {
    console.error('Failed to fetch nodes for globe:', error)
  }
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, { threshold: 0.1 })
  document.querySelectorAll('.fade-in').forEach(el => observer.observe(el))
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
          <h1 class="hero-title">Mycelium&nbsp;Cloud</h1>
          <p class="hero-subtitle">Revolutionary Kubernetes platform that transforms how teams deploy and manage cloud-native applications at scale</p>
        </div>
        <div class="globe-wrapper">
          <FeatureGlobe :width="globeSize" :height="globeSize" :nodes="globeNodes" :labels="globeLabels" />
        </div>
      </div>
      <div class="hero-stats">
        <div class="stat-card fade-in" v-for="stat in gridCards" :key="stat.label">
          <div v-if="isGridStatsLoading" class="d-flex justify-center align-center" style="height: 2.5rem;">
            <v-progress-circular
              indeterminate
              color="primary"
              size="24"
              width="2"
            ></v-progress-circular>
          </div>
          <div v-else class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="home-section section-padding fade-in">
      <div class="container-padding">
        <div class="section-header text-center mb-8">
          <h2 class="section-title">
            Everything You Need to Succeed
          </h2>
          <p class="section-subtitle">
            Powerful tools and features designed for modern cloud-native applications
          </p>
        </div>
        <v-row class="feature-cards-row">
          <v-col cols="12" md="4" v-for="feature in features" :key="feature.title" class="feature-col">
            <div class="home-card card-enhanced fade-in">
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
    <section class="cta-section section-padding fade-in">
      <div class="container-padding">
        <div class="cta-content text-center">
          <h2 class="cta-title">
            Ready to Transform Your Kubernetes Experience?
          </h2>
          <p class="cta-description">
            Join thousands of developers and DevOps engineers who trust Mycelium Cloud for their production workloads.
          </p>
          <v-btn
            v-if="userStore.isLoggedIn"
            variant="outlined"
            color="white"
            size="x-large"
            to="/deploy"
          >
            Deploy Cluster
          </v-btn>
          <v-btn
            v-else
            variant="outlined"
            color="white"
            size="x-large"
            to="/deploy"
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
  position: relative;
  width: 100%;
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
}

.hero-globe-section,
.home-section,
.cta-section {
  background: none;
  box-shadow: none;
  border: none;
  margin: 0;
  width: 100%;
}

.hero-globe-section {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 0;
  margin: 0;
  position: relative;
  z-index: 1;
}

.home-section {
  padding-top: 3rem;
}

.cta-section {
  padding-bottom: 8rem;
}

/* Remove any margin between sections */
.hero-globe-section + .home-section {
  margin-top: 0;
}

/* Features Section */
.home-section {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.section-header {
  margin-bottom: 5rem;
}

.section-title {
  font-size: clamp(2.5rem, 5vw, 3.5rem);
  font-weight: 500;
  margin-bottom: 1.5rem;
  line-height: 1.2;
  color: #fff;
  letter-spacing: -0.5px;
}

.section-subtitle {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  opacity: 0.85;
  max-width: 700px;
  margin: 0 auto;
  font-weight: 400;
}

.home-card {
  text-align: center;
  height: 100%;
}

.feature-cards-row {
  margin: 4rem;
}

.feature-col {
  padding: 1rem;
}

.home-icon {
  margin-bottom: 2.5rem;
}

.home-title {
  font-size: 1.5rem;
  color: #fff;
  margin-bottom: 1.5rem;
  font-weight: 500;
}

.home-description {
  color: #CBD5E1;
  line-height: 1.7;
  opacity: 0.85;
  font-size: 1.1rem;
  font-weight: 400;
}

/* CTA Section */
.cta-content {
  max-width: 700px;
  margin: 0 auto;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3rem;
}

.cta-title {
  font-size: clamp(2.5rem, 6vw, 3.5rem);
  font-weight: 500;
  margin-bottom: 2rem;
  line-height: 1.2;
  color: #fff;
  letter-spacing: -1px;
}

.cta-description {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  margin-bottom: 3rem;
  opacity: 0.92;
  line-height: 1.7;
  font-weight: 400;
}

.cta-btn {
  font-size: 1.25rem;
  padding: 1.3rem 3.5rem;
  border-radius: 1.5rem;
  font-weight: 400;
  box-shadow: 0 4px 24px 0 rgba(59,130,246,0.18);
  background: linear-gradient(90deg, #60a5fa 0%, #38bdf8 100%);
  color: #fff;
  transition: box-shadow 0.2s, transform 0.2s;
}

.cta-btn:hover {
  box-shadow: 0 8px 32px 0 rgba(59,130,246,0.28);
  transform: translateY(-2px) scale(1.04);
}

/* Responsive Design */
@media (max-width: 1100px) {
  .hero-globe-content {
    flex-direction: column;
    gap: 2rem;
    padding: 0 2vw;
    align-items: center;
  }
  .hero-globe-text {
    text-align: center;
    max-width: 100%;
  }
  .globe-wrapper {
    justify-content: center;
    max-width: 90vw;
    width: 60vw;
  }
  .hero-globe-section {
    padding: 2rem 0;
  }
}

@media (max-width: 600px) {
  .globe-wrapper {
    width: 90vw;
    min-width: 0;
  }
}

.hero-globe-content {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 3rem;
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  flex: 1;
  overflow: visible;
  padding: 0 2vw 0 4vw;
  box-sizing: border-box;
  margin-top: 1rem;
}
.hero-globe-text {
  flex: 1 1 420px;
  min-width: 320px;
  max-width: 540px;
  text-align: left;
  z-index: 2;
}
.hero-title {
  font-size: clamp(3rem, 7vw, 5rem);
  font-weight: 500;
  color: #fff;
  margin-bottom: 3rem;
  letter-spacing: -1px;
  line-height: 1.1;
}
.hero-subtitle {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  opacity: 0.95;
  max-width: 500px;
  line-height: 1.5;
  font-weight: 400;
}
.globe-wrapper {
  filter: drop-shadow(0 0 10px #60a5fa11) drop-shadow(0 0 5px #38bdf811);
  width: 40vw;
  aspect-ratio: 1/1;
}
.globe-wrapper::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 800px;
  height: 800px;
  background: radial-gradient(circle, rgba(96, 165, 250, 0.1) 0%, transparent 70%);
  border-radius: 50%;
  z-index: -1;
  animation: pulse 4s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% {
    opacity: 0.3;
    transform: translate(-50%, -50%) scale(1);
  }
  50% {
    opacity: 0.6;
    transform: translate(-50%, -50%) scale(1.1);
  }
}
.hero-stats {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
  opacity: 0.85;
  flex-wrap: wrap;
  margin-bottom: 2rem;
}
.stat-card {
  background: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(96, 165, 250, 0.2);
  border-radius: 1rem;
  padding: 1rem 1.5rem;
  box-shadow: none;
  text-align: center;
  min-width: 140px;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}
.stat-card:hover {
  background: rgba(59, 130, 246, 0.1);
  border-color: rgba(96, 165, 250, 0.4);
  transform: translateY(-2px);
  opacity: 1;
}
.stat-value {
  font-size: 1.4rem;
  color: #60a5fa;
  margin-bottom: 0.3rem;
  line-height: 1;
  font-weight: 500;
}
.stat-label {
  font-size: 1.2rem;
  color: #94a3b8;
  line-height: 1.2;
  font-weight: 400;
}

.fade-in {
  opacity: 0;
  transform: translateY(40px);
  transition: opacity 0.7s cubic-bezier(0.4,0,0.2,1), transform 0.7s cubic-bezier(0.4,0,0.2,1);
}
.fade-in.visible {
  opacity: 1;
  transform: none;
}
.footer-glow {
  display: none;
}
@media (max-width: 900px) {
  .hero-stats {
    gap: 1rem;
  }
}
@media (max-width: 600px) {
  .hero-stats {
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }
}
</style>
