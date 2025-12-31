<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">Users</h1>
      <p class="text-body-1 mt-2 opacity-70">All platform users</p>
    </div>

    <v-data-table-server
      v-model:items-per-page="limit"
      v-model:page="page"
      :headers="[
        { title: 'id', key: 'id', align: 'center', sortable: false },
        { title: 'name', key: 'name', align: 'center', sortable: false },
        { title: 'email', key: 'email', align: 'center', sortable: false },
        { title: 'balance', key: 'balance', align: 'center', sortable: false },
        { title: 'created at', key: 'createdAt', align: 'center', sortable: false },
        { title: 'actions', key: 'actions', align: 'center', sortable: false },
      ]"
      :items="users"
      :items-length="users.length"
      :loading="isLoading"
    >
      <template #item="{ item }">
        <UserRow :user="item" />
      </template>
    </v-data-table-server>

    <v-dialog
      :model-value="credit.isRevealed.value"
      max-width="600"
      scrollable
      @update:model-value="credit.cancel()"
    >
      <UserCreditDialogCard
        :user="credit.data!.value"
        @confirm="credit.confirm($event)"
        @cancel="credit.cancel()"
      />
    </v-dialog>

    <v-dialog
      :model-value="drain.isRevealed.value"
      max-width="600"
      scrollable
      @update:model-value="drain.cancel()"
    >
      <UserDrainDialogCard
        :user="drain.data!.value"
        @confirm="drain.confirm($event)"
        @cancel="drain.cancel()"
      />
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { ServicesUserWithUSDBalance } from "../../generated/api"

const api = useApi()

const limit = ref(10)
const page = ref(1)

const { state: users, isLoading } = useAsyncState(async () => {
  const { data } = await api.admin.getAllUsers()
  return data.data?.users ?? []
}, [])

const credit = useDialog<ServicesUserWithUSDBalance, { amount: number; memo: string }>()
const drain = useDialog<ServicesUserWithUSDBalance>()
// const remove = useDialog<ServicesUserWithUSDBalance>()

provide(UserDialogCtxKey, {
  async credit(user: ServicesUserWithUSDBalance) {
    const { data } = await credit.reveal(user)
    return data
  },

  async drain(user) {
    const { isCanceled } = await drain.reveal(user)
    return !isCanceled
  },

  async remove() {
    return true
  },
})
</script>
