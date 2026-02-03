<template>
  <v-container v-if="md" class="mb-16">
    <div class="my-16">
      <h1 class="text-h2 font-weight-bold text-center mb-4" v-text="md.title" />
      <p
        class="text-body-1 text-center text-primary mx-auto"
        :style="{ maxWidth: '600px' }"
        v-text="`Last Updated: ${md.updatedAt}`"
      />
    </div>

    <v-card>
      <div v-html="md.html" />
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import fm from "front-matter"

definePageMeta({ middleware: "public" })

const { state: md } = useAsyncState(async () => {
  const res = await fetch("/legal/privacy-policy.md")
  const text = await res.text()
  const { attributes, body } = fm(text) as { attributes: Record<string, string>, body: string }
  return { title: attributes.title, updatedAt: attributes.updatedAt, html: marked.parse(body, { renderer }) }
}, null)
</script>
