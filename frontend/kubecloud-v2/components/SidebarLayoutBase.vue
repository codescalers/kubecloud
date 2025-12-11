<template>
  <v-container class="d-flex align-start py-0">
    <v-list>
      <slot name="sidebar"/>
    </v-list>

    <div>
      <slot />
    </div>
  </v-container>
</template>
