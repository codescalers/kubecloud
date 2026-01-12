<template>
  <div>
    slug {{ $route.params.ids }}
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: (to) => {
    const ids = to.params.ids as string[]
    if (ids.length !== 1) {
      return abortNavigation()
    }
  },
})
// const { slug } = useRoute()
</script>
