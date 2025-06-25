<template>
  <div class="pricing-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title">Transparent Pricing</h1>
          <p class="hero-description">
            Farmer-driven pricing for dedicated Kubernetes nodes with no hidden fees.
          </p>
        </div>
      </div>
    </section>

    <!-- Pricing Content -->
    <section class="pricing-content section-padding">
      <div class="container-padding">
        <v-row justify="center">
          <v-col cols="12" lg="8">
              <h2 class="card-title">
                How Pricing Works
              </h2>
              <p class="card-description">
                Node prices are set by independent farmers. 80% of the price goes to the farmer, 20% to the ThreeFold treasury. No hidden fees.
              </p>

              <v-divider class="my-6" color="primary" />

              <h3 class="section-title">
                Sample Pricing
              </h3>
              <div class="table-container pricing-card card-enhanced fade-in">
                <v-table class="pricing-table">
                  <thead>
                    <tr>
                      <th>Node Type</th>
                      <th>CPU</th>
                      <th>RAM</th>
                      <th>GPU</th>
                      <th>Price (USD/month)</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>Standard</td>
                      <td>8</td>
                      <td>32 GB</td>
                      <td>-</td>
                      <td class="price">$180</td>
                    </tr>
                    <tr>
                      <td>High-Memory</td>
                      <td>16</td>
                      <td>64 GB</td>
                      <td>-</td>
                      <td class="price">$350</td>
                    </tr>
                    <tr>
                      <td>GPU</td>
                      <td>16</td>
                      <td>64 GB</td>
                      <td>NVIDIA A100</td>
                      <td class="price">$450</td>
                    </tr>
                  </tbody>
                </v-table>
              </div>
          </v-col>
        </v-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'

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
.pricing-view {
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
  color: var(--kubecloud-text-secondary);
  line-height: 1.7;
  opacity: 0.8;
}

/* Pricing Content */
.pricing-content {
  position: relative;
  z-index: 2;
  padding: 4rem 0;
}

.pricing-card {
  padding: 3rem;
  text-align: center;
}

.card-title {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 700;
  margin-bottom: 1.5rem;
  line-height: 1.3;
}

.card-description {
  font-size: 1.125rem;
  color: var(--kubecloud-text-secondary);
  line-height: 1.7;
  margin-bottom: 2rem;
  opacity: 0.8;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 2rem;
}

.table-container {
  background: var(--kubecloud-overlay-bg, rgba(15,23,42,0.3));
  border-radius: 12px;
  padding: 1.5rem;
  border: 1px solid var(--kubecloud-border, rgba(79,70,229,0.2));
}

.pricing-table {
  background: transparent;
  color: var(--kubecloud-text-primary);
}

.pricing-table th {
  background: var(--kubecloud-primary-bg, rgba(79,70,229,0.1));
  color: #fff;
  font-weight: 600;
  padding: 1rem;
  border-bottom: 1px solid rgba(59, 130, 246, 0.3);
}

.pricing-table td {
  padding: 1rem;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
  color: var(--kubecloud-text-secondary);
}

.price {
  font-weight: 700;
  color: #10B981;
  font-size: 1.125rem;
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

  .pricing-content {
    padding: 2rem 0;
  }

  .pricing-card {
    padding: 2rem;
  }

  .table-container {
    padding: 1rem;
  }

  .pricing-table th,
  .pricing-table td {
    padding: 0.75rem;
    font-size: 0.9rem;
  }
}

@media (max-width: 480px) {
  .hero-section {
    min-height: 40vh;
  }

  .pricing-content {
    padding: 1.5rem 0;
  }

  .pricing-card {
    padding: 1.5rem;
  }

  .pricing-table th,
  .pricing-table td {
    padding: 0.5rem;
    font-size: 0.8rem;
  }
}
</style>
