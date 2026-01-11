<template>
  <div @click="stop()">
    <v-card
      class="md-editor bg-2 elevation-0 rounded-lg"
      :style="{ padding: '0 !important' }"
      :class="{ 'md-editor--focused': focused }"
    >
      <v-card-text class="md-editor-toolbar border-b bg-2">
        <div ref="toolbar" class="d-flex align-center">
          <span class="ql-formats d-flex align-center flex-wrap ga-4">
            <v-select
              v-model="font"
              :items="[
                { text: 'Serif', value: 'serif' },
                { text: 'Monospace', value: 'monospace' },
              ]"
              item-title="text"
              item-value="value"
              density="compact"
              variant="solo"
              width="155px"
              max-width="155px"
              flat
              hide-details
              tabindex="-1"
            >
              <template #item="{ item, props }">
                <v-list-item v-bind="props" :title="undefined">
                  <span :style="{ fontFamily: `${item.value} !important` }" v-text="item.title" />
                </v-list-item>
              </template>
            </v-select>

            <v-select
              v-model="size"
              :items="[
                { text: 'Small', value: '0.75em' },
                { text: 'Normal', value: '1em' },
                { text: 'Large', value: '1.5em' },
                { text: 'Huge', value: '2em' },
              ]"
              return-object
              item-title="text"
              density="compact"
              variant="solo"
              width="155px"
              max-width="155px"
              flat
              hide-details
              tabindex="-1"
            >
              <template #item="{ item, props }">
                <v-list-item v-bind="props" :title="undefined">
                  <span :style="{ fontSize: `${item.value} !important` }" v-text="item.title" />
                </v-list-item>
              </template>
            </v-select>

            <v-btn
              v-for="{ format, icon, action } in formats"
              :key="format"
              class="px-0"
              :class="{ [`ql-${format}`]: !action }"
              :style="{ minWidth: 'auto !important' }"
              variant="plain"
              size="small"
              :ripple="false"
              tabindex="-1"
              @click="action"
            >
              <v-icon :icon="`mdi-${icon}`" size="large" />
            </v-btn>
          </span>
        </div>
      </v-card-text>

      <v-card-text class="overflow-auto" :style="{ height: '180px' }">
        <div class="position-relative h-100">
          <div ref="editor" class="h-100" :style="{ zIndex: 2 }" />
          <p
            v-if="text.length === 0"
            class="md-editor-placeholder position-absolute top-0 left-0 opacity-30"
            :style="{ zIndex: 1 }"
            v-text="label"
          />
        </div>
      </v-card-text>
    </v-card>

    <v-dialog :model-value="isRevealed" max-width="500" scrollable @update:model-value="cancel()">
      <v-form @submit.prevent="confirm($event.target as HTMLFormElement)">
        <v-card :style="{ padding: '0 !important' }">
          <v-card-title class="px-6 py-4">
            <div class="d-flex align-center justify-space-between">
              <h3 class="text-h5 font-weight-bold">
                Add Link
              </h3>
            </div>
          </v-card-title>

          <v-divider />

          <v-card-text>
            <v-text-field
              name="link"
              placeholder="Enter the link"
              variant="outlined"
              prepend-inner-icon="mdi-link"
              hide-details
              autofocus
            />
          </v-card-text>
          <v-divider />

          <v-card-actions class="px-6 py-4 flex-row-reverse justify-start">
            <v-btn variant="text" color="primary" type="submit">
              Add
            </v-btn>
            <v-btn variant="plain" type="button" @click="cancel()">
              Cancel
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-form>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import Quill from "quill"
import "quill/dist/quill.core.css"

defineProps<{ label: string }>()

const toolbar = ref<HTMLDivElement | null>(null)
const editor = ref<HTMLDivElement | null>(null)
const focused = ref(false)
const text = ref("")

let quill: Quill | null = null

const { isRevealed, reveal, cancel, confirm } = useDialog<undefined, HTMLFormElement>()
const formats = markRaw([
  { format: "bold", icon: "format-bold" },
  { format: "italic", icon: "format-italic" },
  { format: "underline", icon: "format-underline" },
  { format: "list", icon: "format-list-bulleted" },
  {
    format: "link",
    icon: "link",
    async action() {
      const { data: form, isCanceled } = await reveal()
      if (!isCanceled && form) {
        const data = new FormData(form)
        const link = data.get("link") as string
        quill?.format("link", link)
      }
    },
  },
  { format: "strike", icon: "format-strikethrough-variant" },
  { format: "blockquote", icon: "format-quote-close" },
])

const font = ref("sans")
watchImmediate(font, f => quill?.format("font", f))

const size = ref({ text: "Normal", value: "1em" })
watchImmediate(size, (v) => {
  const s = v.text.toLowerCase()
  if (s === "normal") {
    return quill?.format("size", false)
  }
  quill?.format("size", s)
})

const { start, stop } = useTimeoutFn(() => (focused.value = false), 150)

onMounted(() => {
  quill = new Quill(editor.value!, {
    modules: {
      toolbar: toolbar.value!,
    },
  })

  quill.editor.scroll.domNode.setAttribute("spellcheck", "false")
  quill.editor.scroll.domNode.onfocus = () => (focused.value = true)
  quill.editor.scroll.domNode.onblur = start
  quill.on("text-change", () => {
    const md = quill!.getSemanticHTML()
    text.value = md === "<p></p>" ? "" : quill!.getSemanticHTML()
  })
})

onUnmounted(() => {
  if (quill) {
    quill.off("text-change")
    quill.editor.scroll.domNode.onfocus = null
    quill.editor.scroll.domNode.onblur = null
    quill.disable()
    quill = null
  }
})
</script>

<style lang="scss">
.md-editor {
  box-sizing: border-box !important;

  &,
  .md-editor-toolbar {
    transition-duration: 0.28s !important;
    transition-property: border-color, box-shadow, opacity, background !important;
    transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1) !important;
  }

  .ql-editor {
    height: 100% !important;
    outline: none !important;
    border: none !important;
    padding: 0 !important;
    margin: 0 !important;
    color: rgba(255, 255, 255, 0.5) !important;
  }

  &.md-editor--focused,
  &:focus,
  &:hover {
    &,
    .md-editor-toolbar {
      border-color: rgba(255, 255, 255, 0.5) !important;
    }
  }

  &.md-editor--focused {
    border-width: 2px !important;

    .md-editor-toolbar {
      border-width: 2px !important;
    }

    .md-editor-placeholder,
    .md-editor-toolbar {
      margin-top: -1px !important;
      margin-left: -1px !important;
    }
  }
}
</style>
