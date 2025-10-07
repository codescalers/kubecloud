<template>
	<v-select
		v-model="model"
		:items="items"
		:label="label"
		:clearable="clearable"
		item-value="nodeId"
		class="node-select"
	>
		<template #item="{ item, index, props: { title, ...rest } }">
			<div>
				<div v-bind="rest" class="node-option-row">
					<NodeInfoRow
						:node="item.raw"
						:get-node-resources="getResources"
						:gpu-icon="gpuIcon"
						:cpu-label="cpuLabel"
					/>
				</div>
				<v-divider v-if="index < items.length - 1" />
			</div>
		</template>
		<template #selection="{ item }">
			<NodeInfoRow
				:node="item.raw"
				:get-node-resources="getResources"
				:gpu-icon="gpuIcon"
				:cpu-label="cpuLabel"
			/>
		</template>
	</v-select>
</template>
<script setup lang="ts">
	import { computed } from 'vue';
	import NodeInfoRow from './NodeInfoRow.vue';
	const props = withDefaults(defineProps<{
		modelValue: number | null,
		items: any[],
		label?: string,
		clearable?: boolean,
		getNodeResources?: (node: any) => { cpu: number; ram: number; storage: number },
		gpuIcon?: string,
		cpuLabel?: string,
	}>(), {
		label: 'Select Node',
		clearable: false,
		gpuIcon: 'mdi-expansion-card',
		cpuLabel: 'vCPU',
	});
	const emit = defineEmits(['update:modelValue']);
	const model = computed({
		get: () => props.modelValue,
		set: (val: number | null) => emit('update:modelValue', val)
	});

	const getResources = (node: any) => props.getNodeResources?.(node) ?? {
		cpu: node?.cpu ?? 0,
		ram: node?.available_ram ?? 0,
		storage: node?.available_storage ?? 0,
	};
</script>
<style scoped>
	.node-option-row {
		margin: .5rem;
		cursor: pointer;
	}
</style>
