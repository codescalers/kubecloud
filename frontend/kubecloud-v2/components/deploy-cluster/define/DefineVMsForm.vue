<template>
  <v-row>
    <v-col cols="12">
      <DefineVMsClusterName v-model="$props.modelValue.name" />
    </v-col>

    <v-col cols="12" md="6">
      <DefineVMsNodes
        icon="mdi-server"
        node-type="Master"
        :ssh-keys="sshKeys"
        :nodes="$props.modelValue.masters"
        @add-node="$props.modelValue.masters.push(createClusterNode({ type: 'master' }))"
        @remove-node="$props.modelValue.masters = $props.modelValue.masters.filter(master => master.id !== $event)"
      />
    </v-col>

    <v-col cols="12" md="6">
      <DefineVMsNodes
        icon="mdi-console"
        node-type="Worker"
        :ssh-keys="sshKeys"
        :nodes="$props.modelValue.workers"
        @add-node="$props.modelValue.workers.push(createClusterNode())"
        @remove-node="$props.modelValue.workers = $props.modelValue.workers.filter(worker => worker.id !== $event)"
      />
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import type { ModelsSSHKey } from "~/generated/api"

defineProps<{ modelValue: ClusterForm, sshKeys: ModelsSSHKey[] }>()
</script>
