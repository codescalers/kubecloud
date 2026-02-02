<template>
  <StickySidebarLayout :cols="9" :sidebar-width="300">
    <template #sidebar>
      <VCard :style="{ padding: '16px !important' }">
        <v-list>
          <v-list-item class="text-body-1 text-accent font-weight-bold">
            Documentation
          </v-list-item>

          <v-list-item
            v-for="doc in docs"
            :key="doc.path"
            :prepend-icon="doc.icon"
            link
            exact
            rounded
            :title="doc.title"
            :to="ROUTES.Docs(doc.path)"
            class="text-accent"
          />

          <div class="mt-4" />

          <v-list-item class="text-body-1 text-accent font-weight-bold">
            Table of Contents
          </v-list-item>

          <v-list-item
            v-for="{ id, content } in activePage?.md.tableOfContent"
            :key="id"
            link
            density="compact"
            rounded
            :href="`#${id}`"
          >
            <v-list-item-title
              class="text-body-2 pl-4 text-no-wrap"
            >
              {{ content }}
            </v-list-item-title>
          </v-list-item>
        </v-list>
        <!-- :to="ROUTES.Docs(doc.path)" -->
      </VCard>
    </template>

    <!-- x {{ route.path }}
    {{ isLoading }}
    {{ docs.map(doc => doc.content) }} -->
    <NuxtPage :page="activePage" />
    <!-- <div :style="{ height: '2000px' }" />
    <NuxtPage /> -->
  </StickySidebarLayout>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })

const route = useRoute()
const { state: docs } = useDocs()

const activePage = computed(() => {
  let activePath = route.path
  if (activePath.endsWith("/")) {
    activePath = activePath.slice(0, -1)
  }
  activePath = activePath.replace(ROUTES.Docs(), "")

  return docs.value.find(doc => doc.path === activePath)
})
</script>

<style lang="scss">
.v-list-item--active {
  color: white !important;

  i {
    color: rgb(var(--v-theme-primary)) !important
  }
}
</style>
