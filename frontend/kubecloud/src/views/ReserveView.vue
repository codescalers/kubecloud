<template>
  <div class="reserve-view">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title kubecloud-gradient kubecloud-glow-blue">
            Reserve Your Node
          </h1>
          <p class="hero-description">
            Choose and reserve your dedicated Kubernetes node from our global network.
          </p>
        </div>
      </div>
    </section>

    <!-- Reservation Content -->
    <section class="reservation-content section-padding">
      <div class="container-padding">
        <v-row>
          <!-- Filter Column -->
          <v-col cols="12" md="3">
            <div class="filter-card card-enhanced fade-in">
              <h3 class="filter-title">Filters</h3>
              
              <!-- CPU Filter -->
              <div class="filter-section">
                <label class="filter-label">CPU Cores</label>
                <v-select
                  v-model="filters.cpu"
                  :items="cpuOptions"
                  variant="outlined"
                  density="compact"
                  hide-details
                  class="filter-select"
                  placeholder="All CPU cores"
                />
              </div>

              <!-- RAM Filter -->
              <div class="filter-section">
                <label class="filter-label">RAM</label>
                <v-select
                  v-model="filters.ram"
                  :items="ramOptions"
                  variant="outlined"
                  density="compact"
                  hide-details
                  class="filter-select"
                  placeholder="All RAM sizes"
                />
              </div>

              <!-- GPU Filter -->
              <div class="filter-section">
                <label class="filter-label">GPU</label>
                <v-select
                  v-model="filters.gpu"
                  :items="gpuOptions"
                  variant="outlined"
                  density="compact"
                  hide-details
                  class="filter-select"
                  placeholder="All GPU types"
                />
              </div>

              <!-- Price Range Filter -->
              <div class="filter-section">
                <label class="filter-label">Price Range</label>
                <v-range-slider
                  v-model="filters.priceRange"
                  :min="0"
                  :max="1000"
                  :step="50"
                  thumb-label="always"
                  class="filter-slider"
                  color="primary"
                />
                <div class="price-range-display">
                  ${{ filters.priceRange[0] }} - ${{ filters.priceRange[1] }}
                </div>
              </div>

              <!-- Location Filter -->
              <div class="filter-section">
                <label class="filter-label">Location</label>
                <v-select
                  v-model="filters.location"
                  :items="locationOptions"
                  variant="outlined"
                  density="compact"
                  hide-details
                  class="filter-select"
                  placeholder="All locations"
                />
              </div>

              <!-- Clear Filters Button -->
              <v-btn
                variant="outlined"
                color="primary"
                class="clear-filters-btn"
                @click="clearFilters"
              >
                Clear Filters
              </v-btn>
            </div>
          </v-col>

          <!-- Nodes Column -->
          <v-col cols="12" md="9">
            <div class="reservation-card card-enhanced fade-in">
              <div class="nodes-header">
                <h2 class="card-title kubecloud-gradient kubecloud-glow-blue">
                  Available Nodes
                </h2>
                <div class="nodes-count">
                  {{ filteredNodes.length }} of {{ availableNodes.length }} nodes
                </div>
              </div>
              <p class="card-description">
                Browse through our available nodes and select the one that best fits your requirements.
              </p>
              
              <v-divider class="my-6" color="primary" />
              
              <div v-if="filteredNodes.length === 0" class="no-results">
                <v-icon size="64" color="primary" class="mb-4">mdi-magnify-close</v-icon>
                <h3>No nodes match your filters</h3>
                <p>Try adjusting your filter criteria to see more options.</p>
                <v-btn
                  color="primary"
                  variant="outlined"
                  @click="clearFilters"
                >
                  Clear All Filters
                </v-btn>
              </div>
              
              <div v-else class="nodes-grid">
                <div v-for="node in filteredNodes" :key="node.id" class="node-card">
                  <div class="node-header">
                    <h3 class="node-title">{{ node.name }}</h3>
                    <div class="node-price">${{ node.price }}/month</div>
                  </div>
                  <div class="node-location" v-if="node.location">
                    <v-icon size="16" class="mr-1">mdi-map-marker</v-icon>
                    {{ node.location }}
                  </div>
                  <div class="node-specs">
                    <div class="spec-item">
                      <span class="spec-label">CPU:</span>
                      <span class="spec-value">{{ node.cpu }} cores</span>
                    </div>
                    <div class="spec-item">
                      <span class="spec-label">RAM:</span>
                      <span class="spec-value">{{ node.ram }} GB</span>
                    </div>
                    <div class="spec-item">
                      <span class="spec-label">Storage:</span>
                      <span class="spec-value">{{ node.storage }} GB</span>
                    </div>
                    <div class="spec-item" v-if="node.gpu">
                      <span class="spec-label">GPU:</span>
                      <span class="spec-value">{{ node.gpu }}</span>
                    </div>
                  </div>
                  <v-btn 
                    color="primary" 
                    variant="elevated" 
                    class="reserve-btn"
                    @click="reserveNode(node.id)"
                  >
                    Reserve Node
                  </v-btn>
                </div>
              </div>
            </div>
          </v-col>
        </v-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const availableNodes = [
  {
    id: 1,
    name: 'Standard Node',
    cpu: 8,
    ram: 32,
    storage: 500,
    price: 180,
    location: 'US East'
  },
  {
    id: 2,
    name: 'High-Memory Node',
    cpu: 16,
    ram: 64,
    storage: 1000,
    price: 350,
    location: 'US West'
  },
  {
    id: 3,
    name: 'GPU Node',
    cpu: 16,
    ram: 64,
    storage: 1000,
    gpu: 'NVIDIA A100',
    price: 450,
    location: 'Europe'
  },
  {
    id: 4,
    name: 'Performance Node',
    cpu: 32,
    ram: 128,
    storage: 2000,
    price: 650,
    location: 'Asia Pacific'
  },
  {
    id: 5,
    name: 'Budget Node',
    cpu: 4,
    ram: 16,
    storage: 250,
    price: 120,
    location: 'US East'
  },
  {
    id: 6,
    name: 'Enterprise Node',
    cpu: 64,
    ram: 256,
    storage: 4000,
    gpu: 'NVIDIA H100',
    price: 1200,
    location: 'US West'
  }
]

// Filter options
const cpuOptions = [
  { title: 'All CPU cores', value: null },
  { title: '4 cores', value: 4 },
  { title: '8 cores', value: 8 },
  { title: '16 cores', value: 16 },
  { title: '32 cores', value: 32 },
  { title: '64 cores', value: 64 }
]

const ramOptions = [
  { title: 'All RAM sizes', value: null },
  { title: '16 GB', value: 16 },
  { title: '32 GB', value: 32 },
  { title: '64 GB', value: 64 },
  { title: '128 GB', value: 128 },
  { title: '256 GB', value: 256 }
]

const gpuOptions = [
  { title: 'All GPU types', value: null },
  { title: 'No GPU', value: 'none' },
  { title: 'NVIDIA A100', value: 'NVIDIA A100' },
  { title: 'NVIDIA H100', value: 'NVIDIA H100' }
]

const locationOptions = [
  { title: 'All locations', value: null },
  { title: 'US East', value: 'US East' },
  { title: 'US West', value: 'US West' },
  { title: 'Europe', value: 'Europe' },
  { title: 'Asia Pacific', value: 'Asia Pacific' }
]

// Filter state
const filters = ref({
  cpu: null,
  ram: null,
  gpu: null,
  priceRange: [0, 1000],
  location: null
})

// Computed filtered nodes
const filteredNodes = computed(() => {
  return availableNodes.filter(node => {
    // CPU filter
    if (filters.value.cpu && node.cpu !== filters.value.cpu) {
      return false
    }
    
    // RAM filter
    if (filters.value.ram && node.ram !== filters.value.ram) {
      return false
    }
    
    // GPU filter
    if (filters.value.gpu) {
      if (filters.value.gpu === 'none' && node.gpu) {
        return false
      }
      if (filters.value.gpu !== 'none' && node.gpu !== filters.value.gpu) {
        return false
      }
    }
    
    // Price range filter
    if (node.price < filters.value.priceRange[0] || node.price > filters.value.priceRange[1]) {
      return false
    }
    
    // Location filter
    if (filters.value.location && node.location !== filters.value.location) {
      return false
    }
    
    return true
  })
})

const clearFilters = () => {
  filters.value = {
    cpu: null,
    ram: null,
    gpu: null,
    priceRange: [0, 1000],
    location: null
  }
}

const reserveNode = (nodeId: number) => {
  // Reservation logic here
  console.log(`Reserving node ${nodeId}`)
}

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
.reserve-view {
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

/* Reservation Content */
.reservation-content {
  position: relative;
  z-index: 2;
  padding: 4rem 0;
}

/* Filter Card */
.filter-card {
  padding: 2rem;
  height: fit-content;
  position: sticky;
  top: 2rem;
}

.filter-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 2rem;
  text-align: center;
}

.filter-section {
  margin-bottom: 1.5rem;
}

.filter-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #CBD5E1;
  margin-bottom: 0.5rem;
}

.filter-select {
  width: 100%;
}

.filter-slider {
  margin-top: 0.5rem;
}

.price-range-display {
  text-align: center;
  font-size: 0.875rem;
  color: #10B981;
  font-weight: 500;
  margin-top: 0.5rem;
}

.clear-filters-btn {
  width: 100%;
  margin-top: 1rem;
}

/* Reservation Card */
.reservation-card {
  padding: 3rem;
}

.nodes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-title {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 700;
  margin: 0;
  line-height: 1.3;
}

.nodes-count {
  font-size: 0.875rem;
  color: #CBD5E1;
  opacity: 0.7;
}

.card-description {
  font-size: 1.125rem;
  color: #CBD5E1;
  line-height: 1.7;
  margin-bottom: 2rem;
  opacity: 0.8;
}

.no-results {
  text-align: center;
  padding: 4rem 2rem;
  color: #CBD5E1;
}

.no-results h3 {
  color: #fff;
  margin-bottom: 1rem;
}

.no-results p {
  margin-bottom: 2rem;
  opacity: 0.8;
}

.nodes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
  margin-top: 2rem;
}

.node-card {
  background: rgba(15, 23, 42, 0.3);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 12px;
  padding: 2rem;
  text-align: left;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.node-card:hover {
  transform: translateY(-4px);
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 8px 32px rgba(59, 130, 246, 0.1);
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.node-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.node-price {
  font-size: 1.125rem;
  font-weight: 700;
  color: #10B981;
}

.node-location {
  font-size: 0.875rem;
  color: #CBD5E1;
  opacity: 0.7;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
}

.node-specs {
  margin-bottom: 2rem;
}

.spec-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.spec-label {
  color: #CBD5E1;
  opacity: 0.7;
}

.spec-value {
  color: #fff;
  font-weight: 500;
}

.reserve-btn {
  width: 100%;
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
  
  .reservation-content {
    padding: 2rem 0;
  }
  
  .filter-card {
    position: static;
    margin-bottom: 2rem;
  }
  
  .reservation-card {
    padding: 2rem;
  }
  
  .nodes-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .nodes-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
  
  .node-card {
    padding: 1.5rem;
  }
}

@media (max-width: 480px) {
  .hero-section {
    min-height: 40vh;
  }
  
  .reservation-content {
    padding: 1.5rem 0;
  }
  
  .filter-card,
  .reservation-card {
    padding: 1.5rem;
  }
  
  .node-card {
    padding: 1rem;
  }
  
  .node-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
}
</style>
