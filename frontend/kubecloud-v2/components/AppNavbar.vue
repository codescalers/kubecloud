<template>
  <v-app-bar>
    <v-container class="d-flex align-center py-0">
      <div class="mr-16">
        <v-img
          src="https://staging.myceliumcloud.tf/assets/logo-DY_xoQWE.png"
          width="150"
          height="40"
        />
      </div>

      <div class="position-relative">
        <div v-for="r in routes" ref="navbarLinkItems" :key="r.path" class="d-inline-block">
          <v-btn
            variant="plain"
            :text="r.title"
            :to="r.path"
            :active="activeRoute == r.path"
            class="navbar-link-item opacity-100"
          />
        </div>

        <span
          class="position-absolute bottom-0 left-0 bg-primary"
          :style="{
            height: '2px',
            width: '1px',
            transition: 'transform 0.5s ease',
            transformOrigin: 'left center',

            transform: `translateX(${activeItem.offset}px) scaleX(${activeItem.width})`,
          }"
        />
      </div>

      <v-spacer />

      <v-btn variant="outlined" class="mr-2" to="/login">Login</v-btn>
      <v-btn variant="flat" color="primary" to="/register">Register</v-btn>
      <!-- <v-menu>
        <template #activator="{ props }">

          <v-btn text="Account" v-bind="props" />
        </template>

        <v-list>
          <v-list-item title="Dashboard" to="/dashboard" />
          <v-list-item title="Logout" @click="console.log('logout')" />
        </v-list>
      </v-menu> -->
      <!-- {{ console.log(activeItem) }} -->
    </v-container>
  </v-app-bar>
</template>

<script setup lang="ts">
const route = useRoute()
const routes = ref([
  { title: "Home", path: "/" },
  { title: "Features", path: "/features" },
  { title: "Docs", path: "/docs" },
  { title: "Use Cases", path: "/use-cases" },
])
const activeRoute = computed(() => {
  const matches = routes.value.filter((current) => route.path.startsWith(current.path))
  return matches[matches.length - 1]?.path ?? ""
})

const navbarLinkItems = ref<HTMLDivElement[]>([])
const activeItem = ref({ offset: 0, width: 0 })
watch(
  () => activeRoute.value,
  (active) => (activeItem.value = getActiveItem(active)),
  { immediate: true }
)

function getActiveItem(active: string) {
  console.log({ active })

  const idx = routes.value.findIndex((r) => r.path.startsWith(active))
  if (idx === -1) {
    console.log({ idx })

    return { offset: 0, width: 0 }
  }

  const el = navbarLinkItems.value[idx]
  if (!el) {
    console.log({ el, idx, navbarLinkItems: [...navbarLinkItems.value] })

    return { offset: 0, width: 0 }
  }

  return { offset: el.offsetLeft, width: el.clientWidth }
}
</script>

<style scoped lang="scss">
.navbar-link-item {
  transition: color 0.3s ease;

  &.v-btn--active,
  &:hover {
    color: rgb(var(--v-theme-primary));
  }
}
/*
.navbar-link-item {
  text-transform: none;
  font-weight: 500;
  font-size: 14px;
  color: #fff;
}*/
</style>
