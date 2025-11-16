<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">Admin Overview</h2>
      <p class="section-subtitle">Monitor platform health and key metrics</p>
    </div>
    <StatsGrid :stats="adminStats" />
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
import StatsGrid from './StatsGrid.vue'

interface AdminStat {
  icon: string;
  color: string;
  value: number | string;
  label: string;
}

const { adminStats } = defineProps<{ adminStats: AdminStat[] }>()
</script>
