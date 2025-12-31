<template>
  <DialogCardLayout title="Credit Balance" icon="mdi-cash-plus" @cancel="$emit('cancel')">
    <template #description>
      <div>
        <span class="text-subtitle-2 opacity-50">Apply credits to user:</span>&nbsp;
        <span
          class="text-subtitle-2 text-primary px-2 py-1 rounded-lg"
          :style="{
            backgroundColor: 'rgba(var(--v-theme-primary), var(--v-border-opacity))',
          }"
          >{{ user?.email ?? "N/A" }}</span
        >
      </div>
    </template>

    <v-form ref="form" @submit.prevent="submit()">
      <v-row>
        <v-col cols="12">
          <div class="d-flex justify-space-between align-center">
            <p class="text-h6">Credit Details</p>
            <p class="text-subtitle-2 opacity-50">Apply credits to user accounts</p>
          </div>
        </v-col>

        <v-col cols="12">
          <v-text-field
            label="Amount"
            type="number"
            name="amount"
            variant="outlined"
            min="0"
            step="0.01"
            prepend-inner-icon="mdi-currency-usd"
            autofocus
            :rules="[
              (v) => !!v || 'Amount is required',
              (v) => !v.includes('e') || 'Amount is invalid',
              (v) => v > 0 || 'Amount must be greater than 0',
              (v) => v < 10000 || 'Amount must be less than 10,000',
              (v) => !isNaN(parseFloat(v)) || 'Amount must be a number',
              (v) =>
                !v.includes('.') ||
                (v.includes('.') && v.split('.')[1].length <= 2) ||
                'Amount can only have 2 decimal places',
            ]"
          >
            <template #append-inner>
              <p>USD</p>
            </template>
          </v-text-field>
        </v-col>

        <v-col cols="12">
          <v-textarea
            rows="4"
            label="Reason / Memo"
            name="memo"
            variant="outlined"
            prepend-inner-icon="mdi-file-document-outline"
            no-resize
            counter="255"
            persistent-counter
            :rules="[
              (v) => !!v || 'Reason is required',
              (v) => v.length >= 3 || 'Reason must be at least 3 characters',
              (v) => v.length <= 255 || 'Reason must be less than 255 characters',
            ]"
          />
        </v-col>

        <v-col cols="12">
          <v-btn
            type="submit"
            block
            size="x-large"
            class="btn-form"
            text="Apply Credits"
            prepend-icon="mdi-cash-plus"
            variant="outlined"
            :disabled="!form?.isValid"
          />
        </v-col>
      </v-row>
    </v-form>
  </DialogCardLayout>
</template>

<script setup lang="ts">
import type { VForm } from "vuetify/components/VForm"
import type { ServicesUserWithUSDBalance } from "../generated/api"

defineProps<{ user?: ServicesUserWithUSDBalance }>()
const emit = defineEmits<{
  (e: "confirm", event: { amount: number; memo: string }): void
  (e: "cancel"): void
}>()

const form = ref<VForm>()
function submit() {
  const f = new FormData(form.value!.$el as HTMLFormElement)
  emit("confirm", { amount: parseFloat(f.get("amount") as string), memo: f.get("memo") as string })
}
</script>
