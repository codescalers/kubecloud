<template>
  <div class="docs-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title kubecloud-gradient kubecloud-glow-blue">
            Documentation
          </h1>
          <p class="hero-description">
            Everything you need to know about KubeCloud and how to get started.
          </p>
        </div>
      </div>
    </section>

    <!-- Documentation Content -->
    <section class="docs-content section-padding">
      <div class="container-padding">
        <v-row>
          <v-col cols="12" md="6" lg="3" v-for="doc in docs" :key="doc.title">
            <div class="doc-card card-enhanced fade-in">
              <div class="doc-icon">
                <v-icon :icon="doc.icon" size="48" color="primary"></v-icon>
              </div>
              <h3 class="doc-title">{{ doc.title }}</h3>
              <p class="doc-description">{{ doc.description }}</p>
              <v-btn 
                color="primary" 
                variant="outlined" 
                :href="doc.link" 
                target="_blank"
                class="doc-link"
              >
                Read More
              </v-btn>
            </div>
          </v-col>
        </v-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'

const docs = [
  {
    icon: 'mdi-api',
    title: 'API Reference',
    description: 'Complete API documentation with examples and authentication details.',
    link: '#'
  },
  {
    icon: 'mdi-console',
    title: 'CLI Guide',
    description: 'Command-line interface documentation for advanced users and automation.',
    link: '#'
  },
  {
    icon: 'mdi-help-circle',
    title: 'Platform FAQ',
    description: 'Frequently asked questions and troubleshooting guides.',
    link: '#'
  },
  {
    icon: 'mdi-book-open',
    title: 'Tutorials',
    description: 'Step-by-step tutorials for common use cases and scenarios.',
    link: '#'
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
.docs-view {
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

/* Documentation Content */
.docs-content {
  position: relative;
  z-index: 2;
  padding: 4rem 0;
}

.doc-card {
  height: 100%;
  padding: 2.5rem;
  text-align: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.doc-card:hover {
  transform: translateY(-8px);
}

.doc-icon {
  margin-bottom: 1.5rem;
}

.doc-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 1rem;
}

.doc-description {
  color: #CBD5E1;
  line-height: 1.6;
  margin-bottom: 2rem;
  opacity: 0.8;
}

.doc-link {
  font-weight: 600;
  text-transform: none;
  letter-spacing: 0.01em;
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
  
  .docs-content {
    padding: 2rem 0;
  }
  
  .doc-card {
    padding: 2rem;
  }
}

@media (max-width: 480px) {
  .hero-section {
    min-height: 40vh;
  }
  
  .docs-content {
    padding: 1.5rem 0;
  }
  
  .doc-card {
    padding: 1.5rem;
  }
}
</style>
