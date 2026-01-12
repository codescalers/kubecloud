<template>
  <StickySidebarLayout :cols="9">
    <template #items>
      <v-list-item link exact title="Getting Started" :to="ROUTES.Docs()" />
      <v-list-item link exact title="Tutorials" :to="ROUTES.Docs.Tutorials()" />
    </template>

    <NuxtPage />
    <div :style="{ height: '2000px' }" />
    <NuxtPage />
  </StickySidebarLayout>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })
</script>
