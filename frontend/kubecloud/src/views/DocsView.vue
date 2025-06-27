<template>
  <div class="docs-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title">Documentation</h1>
          <p class="section-subtitle">Find guides, API references, and tutorials to get the most out of KubeCloud.</p>
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
              <h2 class="card-title">{{ doc.title }}</h2>
              <p class="body-text">{{ doc.description }}</p>
              <v-btn color="white" variant="outlined" :href="doc.link" target="_blank" class="doc-link">
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

.docs-content {
  position: relative;
  z-index: 2;
  padding: 6rem 0 6rem 0;
}

.doc-card {
  height: 100%;
  padding: 2.5rem;
  text-align: center;
  background: rgba(10, 25, 47, 0.85);
  border-radius: 1.5rem;
  box-shadow: 0 4px 20px rgba(96, 165, 250, 0.08);
  color: #CBD5E1;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: none;
}

.doc-card:hover {
  transform: translateY(-10px) scale(1.03);
  box-shadow: 0 8px 32px rgba(96, 165, 250, 0.18);
  border: none;
}

.doc-card * {
  font-size: 1.1rem;
  font-weight: 400;
}

.doc-icon {
  margin-bottom: 2rem;
}

.card-title {
  font-size: 1.5rem;
  font-weight: 500;
  color: #fff;
  margin-bottom: 1rem;
}

.doc-description {
  font-size: 1.1rem;
  color: #60a5fa;
  opacity: 0.85;
  max-width: 700px;
  margin: 0 auto 2rem auto;
  line-height: 1.6;
  font-weight: 400;
}

.section-title {
  font-size: clamp(2.5rem, 5vw, 3.5rem);
  font-weight: 500;
  margin-bottom: 2rem;
  line-height: 1.2;
  color: #fff;
  letter-spacing: -0.5px;
}

.section-subtitle {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  line-height: 1.7;
  opacity: 0.92;
  margin-bottom: 0;
  font-weight: 400;
}

.doc-link {
  font-weight: 500;
  text-transform: none;
  letter-spacing: 0.01em;
}

@media (max-width: 900px) {
  .hero-section {
    padding: 3rem 0 2rem 0;
  }
  .docs-content {
    padding: 3rem 0 3rem 0;
  }
  .doc-card {
    padding: 1.5rem;
  }
}
@media (max-width: 600px) {
  .hero-title {
    font-size: clamp(2rem, 8vw, 3rem);
  }
  .docs-content {
    padding: 2rem 0 2rem 0;
  }
}
</style>
