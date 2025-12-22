<template>
  <v-app>
    <v-layout>
      <app-navbar />

      <v-main>
        <slot />
      </v-main>
      <app-footer />
    </v-layout>
  </v-app>
</template>
