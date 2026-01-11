<template>
  <tr class="text-no-wrap py-2">
    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ invoice.id }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ invoice.user_id }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">${{ toPrecision(invoice.total ?? 0, 2) }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50"> ${{ toPrecision(invoice.tax ?? 0, 2) }} </span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ createdAt }}</span>
    </td>

    <td class="text-center">
      <v-btn
        variant="plain"
        class="border"
        prepend-icon="mdi-eye-outline"
        size="small"
        text="View"
        @click="$emit('view')"
      />

      <!-- <div class="d-flex align-center justify-end ga-4">
        <v-btn
          variant="text"
          class="border"
          prepend-icon="mdi-cash-plus"
          size="small"
          text="Credit Balance"
          :disabled="!user.verified"
          :loading="isCreditLoading"
          @click="onCredit()"
        />

        <v-btn
          variant="text"
          class="border"
          color="warning"
          prepend-icon="mdi-water-remove"
          size="small"
          text="Drain"
          :loading="isDrainLoading"
          @click="onDrain()"
        />

        <v-btn
          variant="text"
          class="border"
          prepend-icon="mdi-trash-can-outline"
          color="error"
          size="small"
          text="Remove"
          :loading="isRemoveLoading"
          @click="onRemove()"
        />
      </div> -->
    </td>
  </tr>
</template>

<script setup lang="ts">
import type { ModelsInvoice } from "~/generated/api"

const props = defineProps<{ invoice: ModelsInvoice }>()
defineEmits<{ (e: "view"): void }>()

const createdAt = useDateFormat(() => props.invoice.created_at, DATE_FORMAT)
</script>
