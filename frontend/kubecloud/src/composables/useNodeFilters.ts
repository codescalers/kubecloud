import { computed, ref, watch } from 'vue';
import type { NormalizedNode } from '../types/normalizedNode';

export interface NodeFilterState {
  cpu: [number, number];
  ram: [number, number];
  gpu: boolean | null;
  priceRange: [number, number];
  location: string | null;
  storage: [number, number];
}

export function useNodeFilters(nodes: () => NormalizedNode[], initialPriceRange: [number, number] = [0, 1000]) {
  // Compute min/max for CPU and RAM
  const cpuMin = computed(() => Math.min(...nodes().map(n => n.cpu).filter(Boolean), 0));
  const cpuMax = computed(() => Math.max(...nodes().map(n => n.cpu).filter(Boolean), 0));
  const ramMin = computed(() => Math.min(...nodes().map(n => Math.round(n.ram)).filter(Boolean), 0));
  const ramMax = computed(() => Math.max(...nodes().map(n => Math.round(n.ram)).filter(Boolean), 0));
  const priceMin = computed(() => Math.min(...nodes().map(n => typeof n.price_usd === 'number' ? n.price_usd : Infinity)));
  const priceMax = computed(() => Math.max(...nodes().map(n => typeof n.price_usd === 'number' ? n.price_usd : 0)));

  // Location options
  const locationOptions = computed(() => {
    const locations = Array.from(
      new Set(
        nodes().map((n) => {
          let loc = n.locationString.split(',')[0].trim()
          return loc.charAt(0).toUpperCase() + loc.slice(1)
        }),
      ),
    ).sort()
    return [
      { title: 'All locations', value: null },
      ...locations.map((loc) => ({ title: loc, value: loc })),
    ]
  })

  // Storage min/max with safe fallback
  const storageValues = computed(() => nodes().map(n => n.storage).filter(v => typeof v === 'number' && !isNaN(v)));
  const storageMin = computed(() => storageValues.value.length ? Math.min(...storageValues.value) : 0);
  const storageMax = computed(() => storageValues.value.length ? Math.max(...storageValues.value) : 10000);

  // Filter state
  const filters = ref<NodeFilterState>({
    cpu: [cpuMin.value, cpuMax.value],
    ram: [ramMin.value, ramMax.value],
    gpu: null,
    priceRange: initialPriceRange,
    location: null,
    storage: [storageMin.value, storageMax.value],
  });

  // Watch for changes in min/max and update filter ranges accordingly
  watch([cpuMin, cpuMax], ([min, max]) => {
    filters.value.cpu = [min, max];
  });
  watch([ramMin, ramMax], ([min, max]) => {
    filters.value.ram = [min, max];
  });

  // Filtering logic
  const filteredNodes = computed(() => {
    return nodes().filter(node => {
      if (node.cpu < filters.value.cpu[0] || node.cpu > filters.value.cpu[1]) return false;
      if (Math.round(node.ram) < filters.value.ram[0] || Math.round(node.ram) > filters.value.ram[1]) return false;
      if (filters.value.gpu && node.gpu === false) return false;
      if (typeof node.price_usd === 'number' && (node.price_usd < filters.value.priceRange[0] || node.price_usd > filters.value.priceRange[1])) return false;
      if (filters.value.location &&
        !node.locationString.toLowerCase().includes(filters.value.location?.toLowerCase() ?? '')
      )
        return false;
      const storageOk = node.storage >= filters.value.storage[0] && node.storage <= filters.value.storage[1]
      return storageOk;
    });
  });

  function clearFilters() {
    filters.value = {
      cpu: [cpuMin.value, cpuMax.value],
      ram: [ramMin.value, ramMax.value],
      gpu: null,
      priceRange: [priceMin.value, priceMax.value],
      location: null,
      storage: [storageMin.value, storageMax.value],
    };
  }

  return {
    filters,
    filteredNodes,
    cpuMin,
    cpuMax,
    ramMin,
    ramMax,
    priceMin,
    priceMax,
    storageMin,
    storageMax,
    locationOptions,
    clearFilters
  };
} 