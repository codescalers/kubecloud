<template>
	<div>
		<div class="node-id">Node {{ node.nodeId }}</div>
		<div class="chip-row">
			<v-chip color="primary" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
				{{ resources.cpu }} {{ cpuLabel }}
			</v-chip>
			<v-chip color="success" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">mdi-memory</v-icon>
				{{ resources.ram }} GB RAM
			</v-chip>
			<v-chip color="info" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
				{{ resources.storage }} GB Disk
			</v-chip>
			<v-chip v-if="node.gpu" color="deep-purple-accent-2" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">{{ gpuIcon }}</v-icon>
				GPU
			</v-chip>
			<v-chip color="secondary" size="x-small" class="mr-1" variant="outlined">
				{{ node.country }}
			</v-chip>
      <v-chip
				v-if="node.rented"
				color="green"
				variant="outlined"
				size="x-small"
			>
				<v-icon size="12" class="mr-1">mdi-lock</v-icon>
				Reserved
			</v-chip>
			<v-chip
				v-else
				color="orange"
				variant="outlined"
				size="x-small"
			>
				<v-icon size="12" class="mr-1">mdi-share-variant</v-icon>
				Shared
			</v-chip>
		</div>
	</div>
</template>
<script setup lang="ts">
	import { computed } from 'vue';
	const props = withDefaults(defineProps<{
		node: any,
		getNodeResources: (node: any) => { cpu: number; ram: number; storage: number },
		gpuIcon?: string,
		cpuLabel?: string,
	}>(), {
		gpuIcon: 'mdi-expansion-card',
		cpuLabel: 'vCPU',
	});
	const resources = computed(() => props.getNodeResources(props.node));
</script>
<style scoped>
	.node-id {
		font-weight: 600;
		margin-right: 1rem;
		margin-bottom: 0.5rem;
	}
	.chip-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
</style>
