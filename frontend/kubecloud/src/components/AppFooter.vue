<template>
  <v-footer class="app-footer mt-auto">
    <v-container>
      <v-row class="py-3">
        <v-col cols="12" md="6" class="d-flex align-center">
          <router-link to="/" class="text-decoration-none">
            <img :src="logo" alt="Mycelium Cloud Logo" width="110">
          </router-link>
        </v-col>
        <v-col cols="12" md="6" class="d-flex align-center justify-md-end">
          <v-btn
            to="/docs"
            variant="text"
            class="text-blue-lighten-3 mx-1"
          >
            Docs
          </v-btn>
          <v-btn
            href="https://github.com/codescalers/kubecloud"
            target="_blank"
            variant="text"
            class="text-blue-lighten-3 mx-1"
          >
            GitHub
          </v-btn>
        </v-col>
      </v-row>

      <v-row class="pb-2">
        <v-col cols="12" class="text-center">
          <span class="text-blue-lighten-3 text-body-2 opacity-80">
            © {{ new Date().getFullYear() }} Mycelium Cloud. All rights reserved.
          </span>
          <span class="text-blue-lighten-3 text-caption opacity-60 ml-4">
            v{{ version }}
          </span>
        </v-col>
      </v-row>
    </v-container>
  </v-footer>
</template>

<script setup lang="ts">
  import logo from '@/assets/logo.png'
  const version = __APP_VERSION__
</script>

<style scoped>
.app-footer {
  background: rgba(10, 25, 47, 0.65);
}
</style>
