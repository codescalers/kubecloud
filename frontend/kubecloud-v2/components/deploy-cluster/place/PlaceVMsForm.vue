<template>
  <v-row>
    <v-col cols="12">
      <PlaceVMsHead v-model="$props.modelValue" />
    </v-col>

    <v-col cols="12">
      <PlaceVMsNodes :cluster="$props.modelValue" />
    </v-col>

    <!-- <v-col cols="6">
      <DefineVMsNodes
        icon="mdi-server"
        node-type="Master"
        :ssh-keys="sshKeys"
        :nodes="$props.modelValue.masters"
        @add-node="$props.modelValue.masters.push(createClusterNode())"
        @remove-node="$props.modelValue.masters = $props.modelValue.masters.filter(master => master.id !== $event)"
      />
    </v-col>

    <v-col cols="6">
      <DefineVMsNodes
        icon="mdi-console"
        node-type="Worker"
        :ssh-keys="sshKeys"
        :nodes="$props.modelValue.workers"
        @add-node="$props.modelValue.workers.push(createClusterNode())"
        @remove-node="$props.modelValue.workers = $props.modelValue.workers.filter(worker => worker.id !== $event)"
      />
    </v-col> -->
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
