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

      <div>
        <v-btn
          v-for="route in routes"
          :key="route.path"
          variant="plain"
          :text="route.title"
          :to="route.path"
          :active="isActive(route.path, $route.path)"
          class="navbar-link-item opacity-100"
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
    </v-container>
  </v-app-bar>
</template>

<script setup lang="ts">
const routes = markRaw([
  { title: "Home", path: "/" },
  { title: "Features", path: "/features" },
  { title: "Docs", path: "/docs" },
  { title: "Use Cases", path: "/use-cases" },
  // { title: "Dashboard", path: "/dashboard" },
])

function isActive(path: string, current: string) {
  if (path === "/" && path === current) {
    return true
  }

  return path !== "/" && current.startsWith(path)
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
