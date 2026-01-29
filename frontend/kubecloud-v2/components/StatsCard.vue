<template>
  <v-card class="text-center">
    <v-avatar
      :icon="icon"
      :color="color"
      variant="tonal"
      :size="50"
      rounded="lg"
    />

    <p class="text-h5 font-weight-bold mt-4 mb-1 text-no-wrap" v-text="currentValue" />
    <p class="text-subtitle-2 text-uppercase text-accent" v-text="title" />
  </v-card>
</template>

<script setup lang="ts">
const props = defineProps<{
  title: string
  icon: string
  color: string
  value: number
  transform?: (value: number) => string | number
}>()

const { remaining } = useCountdown(100, { interval: 15, immediate: true })
const currentValue = computed(() => {
  const p = (100 - remaining.value) / 100

  let v = props.value * p
  if (v !== 1) {
    v = Math.floor(v)
  }

  return props.transform?.(v) ?? v
})
</script>
