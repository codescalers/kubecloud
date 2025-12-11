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

      <v-btn text="Home" to="/" />
      <v-btn text="Features" to="/features" />
      <v-btn text="Docs" to="/docs" :active="$route.path.startsWith('/docs')" />
      <v-btn text="Use Cases" to="/use-cases" />
      <v-btn text="Dashboard" to="/dashboard" :active="$route.path.startsWith('/dashboard')" />

      <v-spacer />

      <v-menu>
        <template #activator="{ props }">
          <v-btn text="Account" v-bind="props" />
        </template>

        <v-list>
          <v-list-item title="Dashboard" to="/dashboard" />
          <v-list-item title="Logout" @click="console.log('logout')" />
        </v-list>
      </v-menu>
    </v-container>
  </v-app-bar>
</template>
