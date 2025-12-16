<template>
  <v-container>
    <v-layout>
      <v-list>
        <v-list-item link exact title="Getting Started" to="/docs" />
        <v-list-item link exact title="Tutorials" to="/docs/tutorials" />
      </v-list>

      <v-main>
        <NuxtPage />
      </v-main>
    </v-layout>
  </v-container>
</template>
