<template>
	<div>
		<div class="node-id">Node {{ node.nodeId }}</div>
		<div class="chip-row">
			<v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
				{{ resources.cpu }} {{ cpuLabel }}
			</v-chip>
			<v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">mdi-memory</v-icon>
				{{ resources.ram }} GB RAM
			</v-chip>
			<v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
				{{ resources.storage }} GB Disk
			</v-chip>
			<v-chip v-if="node.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
				<v-icon size="14" class="mr-1">{{ gpuIcon }}</v-icon>
				GPU
			</v-chip>
			<v-chip color="secondary" text-color="white" size="x-small" class="mr-1" variant="outlined">
				{{ node.country }}
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
		margin-bottom: 2px;
		margin-right: 1rem;
	}
	.chip-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
</style> 