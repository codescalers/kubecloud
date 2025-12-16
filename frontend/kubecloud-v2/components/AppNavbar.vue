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
        <v-btn
          v-for="r in routes"
          :key="r.path"
          ref="navbarLinkItems"
          variant="plain"
          :text="r.title"
          :to="r.path"
          class="navbar-link-item opacity-100"
        />

        <span
          class="position-absolute bottom-0 left-0 bg-primary"
          :style="{
            height: '2px',
            width: '1px',
            transition: 'transform 0.25s ease',
            transformOrigin: 'left center',
            transform: `translateX(${indicatorTransform.offset}px) scaleX(${indicatorTransform.width})`,
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
import type { VBtn } from "vuetify/components/VBtn"

const route = useRoute()
const routes = ref([
  { title: "Home", path: "/" },
  { title: "Features", path: "/features" },
  { title: "Docs", path: "/docs" },
  { title: "Use Cases", path: "/use-cases" },
])

const navbarLinkItems = useTemplateRefsList<VBtn>()
const indicatorTransform = ref({ offset: 0, width: 1 })

watchDebounced(() => route.path, animateIndicatorToActive, { immediate: true, debounce: 100 })
function animateIndicatorToActive() {
  const { offset: currentOffset } = indicatorTransform.value

  const item = navbarLinkItems.value.find((item) => item.$el.classList.contains("v-btn--active"))
  if (!item) {
    indicatorTransform.value = { offset: currentOffset, width: 0 }
    return
  }

  const offset = item.$el.offsetLeft as number
  const width = item.$el.clientWidth as number

  indicatorTransform.value = {
    offset: Math.min(offset, currentOffset),
    width: width + Math.abs(offset - currentOffset),
  }

  setTimeout(() => {
    indicatorTransform.value = { offset, width }
  }, 250)
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
</style>
