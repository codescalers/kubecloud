<template>
  <DefaultLayoutBase>
    <SidebarLayoutBase>
      <template #sidebar>
        <v-list-item link title="Getting Started" to="/docs" />
        <v-list-item link title="Tutorials" to="/docs/tutorials" />
      </template>

      <slot />
    </SidebarLayoutBase>
  </DefaultLayoutBase>
</template>
