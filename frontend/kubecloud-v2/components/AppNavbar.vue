<template>
  <v-app-bar
    height="73"
    class="border-b bg-background"
    :style="{ boxShadow: '0px 2px 4px 0px #3B82F60D' }"
  >
    <v-container class="d-flex align-center py-0">
      <div class="mr-16">
        <NuxtLink to="/">
          <v-img src="~/assets/images/logo.png" width="124" height="33" />
        </NuxtLink>
      </div>

      <div class="position-relative">
        <v-btn
          v-for="r in routes"
          :key="r.path"
          ref="navbarLinkItems"
          size="small"
          variant="plain"
          :text="r.title"
          :to="r.path"
          class="navbar-link-item opacity-70"
        />

        <span
          class="position-absolute left-0 bg-white"
          :style="{
            bottom: '-20px',
            height: '1px',
            width: '1px',
            transition: 'transform 0.25s ease',
            transformOrigin: 'left center',
            transform: `translateX(${indicatorTransform.offset}px) scaleX(${indicatorTransform.width})`,
          }"
        />
      </div>

      <v-spacer />

      <template v-if="!authenticated">
        <v-btn
          v-if="$route.path !== '/login'"
          variant="outlined"
          class="mr-2"
          to="/login"
          text="Sign in"
        />

        <!-- <v-btn
          v-if="$route.path !== '/register'"
          variant="flat"
          color="primary"
          to="/register"
          text="Register"
        /> -->
      </template>

      <v-menu v-else>
        <template #activator="{ props }">
          <v-btn text="Account" v-bind="props" append-icon="mdi-chevron-down" />
        </template>

        <v-list>
          <v-list-item link title="Dashboard" prepend-icon="mdi-view-dashboard" to="/dashboard" />
          <v-list-item link title="Profile" prepend-icon="mdi-account" to="/profile" />
          <v-divider />
          <v-list-item
            link
            title="Logout"
            prepend-icon="mdi-logout"
            class="text-error"
            @click="logout()"
          />
        </v-list>
      </v-menu>
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
const indicatorTransform = ref({ offset: 0, width: 0 })

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

const router = useRouter()
const { accessToken, refreshToken } = useTokens()
const authenticated = computed(() => !!accessToken.value)
function logout() {
  accessToken.value = ""
  refreshToken.value = ""
  router.push("/")
}
</script>

<style scoped lang="scss">
.navbar-link-item {
  transition: opacity 0.3s ease;

  &.v-btn--active,
  &:hover {
    opacity: 1 !important;
  }
}
</style>
