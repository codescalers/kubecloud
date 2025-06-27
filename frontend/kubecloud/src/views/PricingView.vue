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
              <v-table class="pricing-table restyled-glass-table fade-in">
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
  font-size: clamp(2rem, 5vw, 2.8rem);
  font-weight: 500;
  margin-bottom: 2.5rem;
  line-height: 1.1;
  letter-spacing: -1px;
  color: #fff;
  text-align: center;
}

.hero-description {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  font-weight: 400;
  color: #60a5fa;
  line-height: 1.7;
  opacity: 0.92;
  margin-bottom: 0;
}

.pricing-content {
  position: relative;
  z-index: 2;
  padding: 6rem 0 6rem 0;
}

.section-title, .card-title {
  font-size: clamp(1.7rem, 4vw, 2.2rem);
  font-weight: 500;
  margin-bottom: 2rem;
  line-height: 1.2;
  color: #fff;
  letter-spacing: -0.5px;
  text-align: center;
}

.card-description, .section-subtitle {
  font-size: clamp(1.2rem, 2vw, 1.6rem);
  color: #60a5fa;
  opacity: 0.85;
  max-width: 700px;
  margin: 0 auto 2rem auto;
  line-height: 1.6;
  font-weight: 400;
}

.pricing-card, .table-container {
  padding: 3rem;
  text-align: center;
  background: rgba(10, 25, 47, 0.85);
  border-radius: 1.5rem;
  box-shadow: 0 4px 20px rgba(96, 165, 250, 0.08);
  color: #CBD5E1;
  border: 1px solid rgba(96, 165, 250, 0.15);
}

.pricing-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  background: none;
  font-size: 1.05rem;
  box-shadow: none;
}

.pricing-table th {
  background: rgba(96, 165, 250, 0.13);
  color: #fff;
  font-weight: 500;
  padding: 1rem;
  border-bottom: 1.5px solid #334155;
  text-align: center;
  vertical-align: middle;
}

.pricing-table td {
  padding: 1rem;
  border-bottom: 1px solid rgba(59, 130, 246, 0.10);
  color: #CBD5E1;
  text-align: left;
  vertical-align: middle;
}

.pricing-table tr:last-child td {
  border-bottom: none;
}

.pricing-table tr:hover td {
  background: rgba(96, 165, 250, 0.07);
  transition: background 0.2s;
}

.price {
  font-weight: 400;
  color: #10B981;
  font-size: 1.125rem;
}

.pricing-description {
  font-size: 1.1rem;
  font-weight: 400;
  color: #60a5fa;
  margin-bottom: 2.5rem;
  line-height: 1.7;
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
  text-align: center;
}

.sample-pricing-outer {
  background: rgba(10, 25, 47, 0.92);
  border: 1.5px solid #334155;
  border-radius: 1.5rem;
  box-shadow: 0 6px 32px rgba(96, 165, 250, 0.10);
  padding: 2.5rem 1.5rem;
  margin: 2.5rem auto 0 auto;
  max-width: 1100px;
}

/* Responsive Design */
@media (max-width: 900px) {
  .hero-section {
    padding: 3rem 0 2rem 0;
  }
  .pricing-content {
    padding: 3rem 0 3rem 0;
  }
  .pricing-card, .table-container {
    padding: 1.5rem;
  }
}
@media (max-width: 600px) {
  .hero-title {
    font-size: clamp(2rem, 8vw, 3rem);
  }
  .pricing-content {
    padding: 2rem 0 2rem 0;
  }
}

.restyled-glass-table {
  margin: 3.5rem auto 3.5rem auto;
  max-width: 900px;
  border-radius: 0.8rem;
  border: 1px solid rgba(96, 165, 250, 0.15);
  overflow: hidden;
}
.restyled-glass-table th {
  background: rgba(96, 165, 250, 0.13);
  color: #fff;
  font-weight: 500;
  padding: 1.1rem;
  border-bottom: 1.5px solid #334155;
  text-align: center;
}
.restyled-glass-table td {
  padding: 1.1rem;
  border-bottom: 1px solid rgba(59, 130, 246, 0.10);
  color: #CBD5E1;
  text-align: left;
  vertical-align: middle;
}
.restyled-glass-table tr:last-child td {
  border-bottom: none;
}
.restyled-glass-table tr:hover td {
  background: rgba(96, 165, 250, 0.07);
  transition: background 0.2s;
}
@media (max-width: 900px) {
  .restyled-glass-table {
    max-width: 100%;
    font-size: 0.98rem;
    margin: 2.5rem auto 2.5rem auto;
  }
  .restyled-glass-table th, .restyled-glass-table td {
    padding: 0.7rem;
  }
}
@media (max-width: 600px) {
  .restyled-glass-table {
    border-radius: 1rem;
    font-size: 0.92rem;
    margin: 1.5rem auto 1.5rem auto;
  }
  .restyled-glass-table th, .restyled-glass-table td {
    padding: 0.5rem;
  }
}
</style>
