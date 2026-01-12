<template>
  <DialogCardLayout
    title="Invoice Details"
    icon="mdi-invoice-list-outline"
    icon-color="success"
    description="Review invoice details"
    @cancel="$emit('cancel')"
  >
    <template #outer>
      <v-row no-gutters class="border-b">
        <v-col class="pa-8" cols="4">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            Total Amount
          </p>
          <p class="text-h3 font-weight-bold">
            ${{ toPrecision(invoice?.total ?? 0, 3) }}
          </p>
        </v-col>

        <v-col class="pa-8 border-s border-e" cols="4">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            Tax Applied
          </p>
          <p class="text-h3 font-weight-bold">
            ${{ toPrecision(invoice?.tax ?? 0, 3) }}
          </p>
        </v-col>

        <v-col class="pa-8" cols="4">
          <p class="text-subtitle-2 opacity-50 mb-1 text-uppercase">
            created date
          </p>
          <p class="text-h6 font-weight-bold" v-text="createdAtDate" />
          <p class="text-subtitle-2 opacity-50" v-text="createdAtTime" />
        </v-col>
      </v-row>

      <v-data-table
        :headers="[
          { title: 'node id', key: 'nodeId', align: 'center', sortable: false },
          { title: 'contract id', key: 'contractId', align: 'center', sortable: false },
          { title: 'period', key: 'period', align: 'center', sortable: false },
          { title: 'cost', key: 'cost', align: 'center', sortable: false },
        ]"
        :items="invoice?.nodes ?? []"
        hide-default-footer
        :style="{ borderTopLeftRadius: '0 !important', borderTopRightRadius: '0 !important', border: 'none !important' }"
      >
        <template #item="{ item }">
          <tr class="text-center">
            <td class="text-subtitle-2 opacity-70" v-text="item.node_id" />
            <td class="text-subtitle-2 opacity-70" v-text="item.contract_id" />
            <td class="text-subtitle-2 opacity-70" v-text="item.period" />
            <td class="text-subtitle-2 opacity-70" v-text="item.cost" />
          </tr>
        </template>
      </v-data-table>
    </template>

    <template #actions>
      <v-btn
        component="a"
        :variant="downloaded ? 'tonal' : 'text'"
        :prepend-icon="downloaded ? 'mdi-check' : 'mdi-download'"
        :text="downloaded ? `00:0${remaining}` : 'Download'"
        color="success"
        download="invoice.json"
        :href="binary"
        :disabled="downloaded"
        @click="startDownload(3)"
      />
    </template>
  </DialogCardLayout>
</template>

<script setup lang="ts">
import type { ModelsInvoice } from "~/generated/api"

const props = defineProps<{ invoice?: ModelsInvoice }>()
defineEmits<{ (e: "cancel"): void }>()

const createdAtDate = useDateFormat(() => props.invoice?.created_at, "MMM d, YYYY")
const createdAtTime = useDateFormat(() => props.invoice?.created_at, "HH:mm:ss UTC")

const binary = computed(() => {
  const data = props.invoice

  if (!data) {
    return ""
  }

  const jsonString = JSON.stringify(data, null, 2)
  const blob = new Blob([jsonString], { type: "application/json" })

  return URL.createObjectURL(blob)
})

const { isActive: downloaded, start: startDownload, remaining } = useCountdown(3)
</script>
