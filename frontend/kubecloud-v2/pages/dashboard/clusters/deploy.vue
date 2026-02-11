<template>
  <div>
    <DefineStepperLine #="{ completed }">
      <div class="position-relative w-100" :style="{ height: '2px', margin: '-34px -67px 0' }">
        <div
          class="w-100 h-100 position-absolute top-0 left-0"
          :style="{ backgroundColor: 'rgba(var(--v-border-color), var(--v-border-opacity))' }"
        />
        <div
          class="w-100 h-100 position-absolute top-0 left-0"
          :style="{
            backgroundColor: 'rgb(var(--v-theme-success))',
            transformOrigin: 'left',
            transition: 'transform 0.3s ease',
            transform: `scaleX(${completed ? 1 : 0})`,
          }"
        />
      </div>
    </DefineStepperLine>

    <DefineStepperItem #="{ title, step, value, completed }">
      <v-stepper-item
        :color="completed || step > value ? 'success' : 'primary'"
        :editable="step === value && !completed"
        :complete="completed || step > value"
        :value="value"
        :class="{ 'text-primary': step === value && !completed, 'text-success': completed || step > value }"
        :title="title"
      />
    </DefineStepperItem>

    <v-container class="py-16">
      <div class="mb-16">
        <h1 class="text-h2 font-weight-bold text-center mb-4">
          Deploy New Cluster
        </h1>
        <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
          Create and configure your Kubernetes cluster in just a few steps
        </p>
      </div>

      <v-stepper
        v-model="step"
        flat
        bg-color="transparent"
        alt-labels
        #="{ next, prev }"
      >
        <v-stepper-header class="elevation-0">
          <ReuseStepperItem title="Define VMs" :step="step" :value="1" :completed="step > 1" />
          <ReuseStepperLine :completed="step > 1" />
          <ReuseStepperItem title="Place VMs" :step="step" :value="2" :completed="step > 2" />
          <ReuseStepperLine :completed="step > 2" />
          <ReuseStepperItem title="Review" :step="step" :value="3" :completed="step === 3" />
        </v-stepper-header>

        <v-stepper-window>
          <v-stepper-window-item eager :value="1">
            <DefineVMsForm v-model="cluster" :ssh-keys="sshKeys" />
          </v-stepper-window-item>

          <v-stepper-window-item eager :value="2">
            <PlaceVMsForm v-model="cluster" />
          </v-stepper-window-item>

          <v-stepper-window-item eager :value="3">
            {{ cluster }}
          </v-stepper-window-item>
        </v-stepper-window>

        <v-divider class="mt-12 mb-6" />

        <div class="d-flex justify-space-between align-center">
          <v-btn
            prepend-icon="mdi-arrow-left"
            text="Back"
            variant="text"
            :disabled="step === 1"
            @click="() => {
              if (step === 2) {
                $nextTick(() => {
                  cluster.masters = cluster.masters.map(master => ({ ...master, node: null }))
                  cluster.workers = cluster.workers.map(worker => ({ ...worker, node: null }))
                })
              }

              prev()
            }
            "
          />

          <v-btn
            v-if="step !== 3"
            append-icon="mdi-arrow-right"
            text="Continue"
            variant="tonal"
            color="primary"
            :disabled="
              (step === 1 && !defineFormValid)
                || (step === 2 && !placeFormValid)
            "
            @click="next"
          />

          <v-btn
            v-else
            prepend-icon="mdi-rocket-launch"
            text="Deploy Cluster"
            variant="tonal"
            color="success"
            :loading="deploying"
            @click="deployCluster()"
          />
        </div>
      </v-stepper>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { HandlersNodeInput } from "~/generated/api"

const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onBeforeUnmount(drawer.open)

onMounted(container.fluidize)
onBeforeUnmount(container.containerize)

const api = useApi()
const { state: sshKeys } = useAsyncState(async () => {
  const { data } = await api.users.listSshKeys()
  return data.data ?? []
}, [], { immediate: $meta.client })

const [DefineStepperLine, ReuseStepperLine] = createReusableTemplate({
  props: { completed: Boolean },
})

const [DefineStepperItem, ReuseStepperItem] = createReusableTemplate({
  props: { title: String, step: Number, value: Number, completed: Boolean },
})

const cluster = ref(createClusterForm())

const step = ref(1)
const defineFormValid = computed(() => {
  const { name, masters, workers } = cluster.value
  return name.length >= 3 && masters.concat(workers).every(isValidClusterNode)
})
const placeFormValid = computed(() => {
  const { masters, workers } = cluster.value
  return masters.every(v => v.node && v.node.valid) && workers.every(v => v.node && v.node.valid)
})

async function toNodeInput(clusterNode: ClusterNode): Promise<HandlersNodeInput> {
  const keys = sshKeys.value
  const SSH_KEY = clusterNode.sshKeys.map(i => keys[i]!.public_key).join("\n")
  const root_size = 5 * 1024

  const input: HandlersNodeInput = {
    name: clusterNode.name,
    type: clusterNode.type === "worker" ? "worker" : "master",
    node_id: clusterNode.node!.id,
    cpu: clusterNode.cpu,
    memory: clusterNode.memory * 1024,
    root_size,
    data_disks: [clusterNode.disk * 1024 - root_size],
    env_vars: {
      SSH_KEY,
      K3S_TOKEN: cluster.value.token,
    },
  }

  if (clusterNode.useFullNodeCapabilities) {
    const { nodeId, total_resources, used_resources } = clusterNode.node!.raw

    input.cpu = total_resources!.cru!
    input.memory = Math.floor(((total_resources!.mru! - used_resources!.mru!) * 0.99) / 1024 ** 3) * 1024

    const pools = await api.nodes.getNodeStoragePool(nodeId!.toString()).then(v => v.data.data?.pools ?? [])
    if (pools.length === 0) {
      throw new Error("No storage pool found")
    }

    input.data_disks = pools.map((pool, i) => {
      const size = Math.floor(pool.free! * 0.985 / 1024 ** 3)
      return size * 1024 - (i === 0 ? root_size : 0)
    })
  }

  return input
}

const { execute: deployCluster, isLoading: deploying } = useAsyncState(async () => {
  const { name, masters, workers } = cluster.value
  const nodes = await Promise.all(masters.concat(workers).map(toNodeInput))
  /* const { data } =  */await api.deployments.deploymentsPost({
    name,
    token: cluster.value.token,
    nodes,
  })
}, null, { immediate: false })
</script>
