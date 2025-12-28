<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-3">
      <p class="text-subtitle-1 font-weight-bold" v-text="title" />
      <v-chip
        :color="color"
        size="small"
        :class="`rounded-lg border border-${color} font-weight-bold`"
        :style="{ '--v-border-opacity': 0.5 }"
        :text="`${modelValue?.map(normalizer).join(' - ') ?? 'N/A'} ${unit}`"
      />
    </div>
    <v-range-slider
      :model-value="modelValue"
      thumb-size="13"
      thumb-color="white"
      track-size="1"
      hide-details
      :color="color"
      :min="range[0]"
      :max="range[1]"
      :step="step"
      @update:model-value="$emit('update:model-value', $event as [number, number])"
    />
    <div class="d-flex justify-space-between align-center text-body-2 opacity-70">
      <p v-text="normalizer(range[0])" />
      <p v-text="normalizer(range[1])" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps({
  title: { type: String, required: true },
  modelValue: { type: [[], null] as PropType<[number, number]> | null, default: null },
  range: { type: [] as PropType<[number, number]>, required: true },
  unit: { type: String, required: true },
  step: { type: [String, Number] as PropType<string | number>, default: 1 },
  color: { type: String, default: "primary" },
  normalizer: {
    type: Function as PropType<(value: number) => string | number>,
    default: (v: number | string) => v,
  },
})

const emit = defineEmits<{ (e: "update:model-value", value: [number, number]): void }>()

onMounted(() => {
  const { modelValue, range } = props

  if (!modelValue) {
    return emit("update:model-value", props.range)
  }

  emit("update:model-value", [
    modelValue[0] > range[0] && modelValue[0] < range[1] ? modelValue[0] : range[0],
    modelValue[1] > range[0] && modelValue[1] < range[1] ? modelValue[1] : range[1],
  ])
})
</script>
