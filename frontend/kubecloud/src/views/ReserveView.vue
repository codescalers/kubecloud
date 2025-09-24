<template>
  <div class="reserve-view">
    <!-- Back Button Section -->
    <section class="back-button-section mx-7">
      <div class="container-padding">
        <v-btn
          color="primary"
          variant="outlined"
          @click="goBack"
          prepend-icon="mdi-arrow-left"
          class="back-btn"
        >
          Back
        </v-btn>
      </div>
    </section>
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content container-padding">
        <div class="hero-text text-center">
          <h1 class="hero-title">
            Reserve Your Node
          </h1>
          <p class="hero-description">
            Choose and reserve your dedicated Kubernetes node from our global network.
          </p>
        </div>
      </div>
    </section>


    <!-- Authentication Banner for Non-Authenticated Users -->
    <section v-if="!isAuthenticated" class="auth-banner section-padding">
      <div class="container-padding">
        <v-alert
          type="info"
          variant="tonal"
          class="auth-alert"
        >
          <template v-slot:prepend>
            <v-icon size="24">mdi-information</v-icon>
          </template>
          <div class="auth-alert-content">
            <h3>Ready to reserve your nodes?</h3>
            <p>Sign in to your account to start reserving nodes and deploying your applications.</p>
            <div class="auth-alert-actions">
              <v-btn
                color="primary"
                variant="elevated"
                @click="router.push('/sign-in')"
                class="mr-3"
              >
                Sign In
              </v-btn>
              <v-btn
                color="secondary"
                variant="outlined"
                @click="router.push('/sign-up')"
              >
                Create Account
              </v-btn>
            </div>
          </div>
        </v-alert>
      </div>
    </section>

    <!-- Reservation Content -->
    <section class="reservation-content section-padding">
      <div class="container-padding">
        <v-row>
          <!-- Filter Column -->
          <v-col cols="12" md="3">
            <NodeFilterPanel
              v-model="filters"
              :cpuMin="cpuMin"
              :cpuMax="cpuMax"
              :ramMin="ramMin"
              :ramMax="ramMax"
              :priceMin="priceMin"
              :priceMax="priceMax"
              :storageMin="storageMin"
              :storageMax="storageMax"
              :locationOptions="locationOptions"
              @clear="clearFilters"
            />
          </v-col>

          <!-- Nodes Column -->
          <v-col cols="12" md="9">
            <v-card class="reservation-card">
              <div class="nodes-header" style="display: flex; align-items: center; justify-content: space-between; gap: 1rem;">
                <div style="display: flex; align-items: center; gap: 1.2rem;">
                  <h2 class="card-title">
                    Available Nodes
                  </h2>
                  <div class="nodes-count">
                    {{ filteredNodes.length }}  of {{ nodes.length }} nodes
                  </div>
                </div>
                <v-btn
                  color="primary"
                  variant="outlined"
                  :disabled="loading"
                  @click="fetchNodes(nodeFilters)"
                  prepend-icon="mdi-refresh"
                  class="refresh-btn"
                  style="min-width: 120px;"
                >
                  Refresh
                </v-btn>
              </div>
              <p class="card-description">
                Browse through our available nodes and select the one that best fits your requirements.
              </p>

              <v-divider class="my-6" color="primary" />

              <div v-if="loading" class="loading-section">
                <v-skeleton-loader type="card, card, card, card" :loading="loading" class="w-100" />
                <p class="loading-text">Loading available nodes...</p>
              </div>
              <template v-else>
                <div v-if="paginatedNodes.length === 0" class="no-results">
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
                <div v-else>
                  <v-row dense align="stretch" style="display: flex;">
                    <v-col
                      v-for="node in paginatedNodes"
                      :key="node.nodeId"
                      class="node-col"
                    >
                      <NodeCard
                        :node="node"
                        :isAuthenticated="isAuthenticated"
                        :loading="reservingNodeId === node.nodeId"
                        :disabled="reservingNodeId === node.nodeId"
                        @action="handleNodeAction(node, $event)"
                        @signin="handleSignIn"
                        tabindex="0"
                        aria-label="Node card"
                      />
                    </v-col>
                  </v-row>
                </div>
                <v-pagination
                  v-model="currentPage"
                  :length="totalPages"
                  color="primary"
                  class="mt-6"
                />
              </template>
            </v-card>
          </v-col>
        </v-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useNodes, type NodeFilters } from '../composables/useNodes'
import { userService } from '../utils/userService'
import { useUserStore } from '../stores/user'
import { useNotificationStore } from '../stores/notifications'
import { useNormalizedNodes } from '../composables/useNormalizedNodes'
import { useNodeFilters } from '../composables/useNodeFilters'
import NodeFilterPanel from '../components/NodeFilterPanel.vue'
import NodeCard from '../components/NodeCard.vue'
import type { NormalizedNode } from "@/types/normalizedNode"

const router = useRouter()
const userStore = useUserStore()
const isAuthenticated = computed(() => userStore.isLoggedIn)

const { nodes, loading, fetchNodes } = useNodes()
const normalizedNodes = useNormalizedNodes(() => nodes.value)
const {
  filters,
  filteredNodes,
  cpuMin, cpuMax,
  ramMin, ramMax,
  priceMin, priceMax,
  storageMin, storageMax,
  locationOptions,
  clearFilters
} = useNodeFilters(() => normalizedNodes.value)

const reservingNodeId = ref<number | null>(null)
const reservedNodeIds = ref(new Set<number>())
const nodeFilters = <NodeFilters>{rentable: true}
onMounted(() => {
  fetchNodes(nodeFilters)
  // Add scroll animation observer
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

const reserveNode = async (node: NormalizedNode) => {
  if (!isAuthenticated.value) {
    router.push('/sign-in')
    return
  }
  reservingNodeId.value = node.nodeId
  try {
    // Fetch user balance
    const { balance } = await userService.fetchBalance()
    // Calculate daily price (assuming monthly price divided by 30)
    const basePrice = Number(node.discount_price ?? node.price_usd ?? 0)
    const extraFee = Number(node.extraFee ?? 0) / 1000
    const dailyPrice = (basePrice + extraFee) / 30

    if (balance < dailyPrice) {
      useNotificationStore().error(
        'Insufficient Balance',
        `Your current balance is not enough to reserve this node. You need to have at least $${dailyPrice.toFixed(2)}. Please add funds to continue.`
      )
      // Redirect to add funds card in dashboard
      localStorage.setItem('dashboard-section', 'add-funds')
      router.push('/dashboard')
      return
    }

    await userService.reserveNode(node.nodeId)
    reservedNodeIds.value.add(node.nodeId) // Optimistically remove from UI
    fetchNodes(nodeFilters)
  } catch (err) {
    console.error(err)
    reservedNodeIds.value.delete(node.nodeId)
  } finally {
    reservingNodeId.value = null
  }
}

const handleSignIn = () => {
  router.push('/sign-in')
}

const goBack = () => {
  router.go(-1)
}

// Pagination logic
const currentPage = ref(1)
const pageSize = 8

const totalPages = computed(() => Math.max(1, Math.ceil(filteredNodes.value.length / pageSize)))
const paginatedNodes = computed(() =>
  filteredNodes.value.slice((currentPage.value - 1) * pageSize, currentPage.value * pageSize).filter(n => !reservedNodeIds.value.has(n.nodeId))
)

function handleNodeAction(node: NormalizedNode, payload: { nodeId: number; action: string }) {
  if (payload.action === 'reserve') {
    reserveNode(node);
  }
}

watch(filteredNodes, () => {
  if (currentPage.value > totalPages.value) {
    currentPage.value = totalPages.value;
  }
  if (currentPage.value < 1) {
    currentPage.value = 1;
  }
})
</script>

<style scoped>
.reserve-view {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  margin-top: 5rem;
}

/* Hero Section */
.hero-section {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
  padding: 2rem 0;
}

.back-button-section {
  padding: 1rem 0;
}

/* Auth Banner */
.auth-banner {
  background: rgba(59, 130, 246, 0.05);
  border-top: 1px solid rgba(59, 130, 246, 0.1);
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
}

.auth-alert {
  background: rgba(59, 130, 246, 0.1) !important;
  border: 1px solid rgba(59, 130, 246, 0.2) !important;
  border-radius: 1rem !important;
}

.auth-alert-content h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1E40AF;
  margin-bottom: 0.5rem;
}

.auth-alert-content p {
  color: #374151;
  margin-bottom: 1rem;
}

.auth-alert-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
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
  font-weight: 400;
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
  padding: 2rem;
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
  border-radius: var(--radius-xl);
}

.nodes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-title {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 400;
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

.loading-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  padding: 3rem 0;
  color: #CBD5E1;
}
.loading-section .v-progress-circular {
  margin-bottom: 1.5rem;
}
.loading-text {
  font-size: 1.15rem;
  color: #CBD5E1;
  opacity: 0.85;
  margin-top: 0.5rem;
}

.node-col {
  flex: 1 1 250px; /* Allow growing and shrinking with a basis of 250px */
  min-width: 250px; /* Enforce the minimum width */
  max-width: 400px;
}

.range-display {
  text-align: center;
  font-size: 0.95rem;
  color: #10B981;
  font-weight: 500;
  margin-top: 0.25rem;
}
</style>
