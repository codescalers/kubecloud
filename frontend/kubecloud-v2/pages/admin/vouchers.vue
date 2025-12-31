<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">Voucher Management</h1>
      <p class="text-body-1 mt-2 opacity-70">Generate and manage platform vouchers</p>
    </div>

    <v-card :style="{ padding: '0 !important' }" class="mb-8">
      <v-card-title class="pa-8 d-flex align-center justify-space-between border-b mb-8">
        <div>
          <div class="d-flex ga-2 align-baseline">
            <v-icon icon="mdi-gift" size="small" color="primary" />
            <span class="text-h6 font-weight-bold">Voucher Management</span>
          </div>
        </div>
      </v-card-title>

      <v-card-text>
        <v-form>
          <v-row>
            <v-col cols="4">
              <v-text-field
                label="Voucher Amount (USD)"
                variant="outlined"
                prepend-inner-icon="mdi-currency-usd"
                placeholder="Enter amount"
              />
            </v-col>
            <v-col cols="4">
              <v-text-field
                label="Number of Vouchers"
                variant="outlined"
                prepend-inner-icon="mdi-pound"
                placeholder="Enter number of vouchers"
              />
            </v-col>
            <v-col cols="4">
              <v-text-field
                label="Expiry (Days)"
                variant="outlined"
                prepend-inner-icon="mdi-calendar-outline"
                placeholder="Enter expiry (days)"
              />
            </v-col>

            <v-col cols="12" class="pt-0 d-flex justify-end">
              <v-btn
                type="submit"
                class="btn-form"
                text="Generate Vouchers"
                prepend-icon="mdi-creation"
                variant="outlined"
              />
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
    </v-card>

    <v-data-table-server
      v-model:items-per-page="limit"
      v-model:page="page"
      :headers="[
        { title: 'Voucher', key: 'code', align: 'center', sortable: false },
        { title: 'Amount', key: 'value', align: 'center', sortable: false },
        { title: 'Redeemed', key: 'redeemed', align: 'center', sortable: false },
        { title: 'Created At', key: 'created_at', align: 'center', sortable: false },
        { title: 'Expiry', key: 'expires_at', align: 'center', sortable: false },
      ]"
      :items="vouchers"
      :items-length="state.length"
      :loading="isLoading"
    >
      <template #[`item.created_at`]="{ item }">
        {{ useDateFormat(item.created_at, "DD/MM/YYYY, HH:mm") }}
      </template>

      <template #[`item.expires_at`]="{ item }">
        {{ useDateFormat(item.expires_at, "DD/MM/YYYY, HH:mm") }}
      </template>
    </v-data-table-server>
  </div>
</template>

<script setup lang="ts">
import type { ModelsVoucher } from "../../generated/api"

const limit = ref(10)
const page = ref(1)

const api = useApi()
const { state, isLoading } = useAsyncState(
  async () => {
    const { data } = await api.admin.listVouchers()
    return (data.data as unknown as { vouchers: ModelsVoucher[] }).vouchers ?? []
  },
  [],
  { resetOnExecute: false }
)

const vouchers = computed(() => {
  const v = (page.value - 1) * limit.value
  return state.value.slice(v, v + limit.value)
})
</script>
