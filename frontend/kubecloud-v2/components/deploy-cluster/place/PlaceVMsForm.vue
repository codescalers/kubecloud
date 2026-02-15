<template>
  <v-row>
    <v-col cols="12">
      <PlaceVMsHead
        v-model="$props.modelValue"
      />
    </v-col>

    <v-col cols="12">
      <PlaceVMsNodes :cluster="$props.modelValue" />
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
defineProps<{ modelValue: ClusterForm }>()

const loadingNode = ref<number | undefined>()

provide(NodePickCtxKey, {
  loadingNode,
  setLoadingNode: (nodeId?: number) => (loadingNode.value = nodeId),
})
</script>
