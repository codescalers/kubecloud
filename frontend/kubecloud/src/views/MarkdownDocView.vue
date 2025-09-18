<template>
  <div class="docs-page">
    <section class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">{{ title }}</h1>
        <p class="section-subtitle">Mycelium Cloud Documentation</p>
      </div>
      <div class="back-actions container-padding">
        <v-btn color="white" variant="outlined" :to="'/docs'" class="back-btn">
          <v-icon icon="mdi-arrow-left" class="mr-2" />
          Back
        </v-btn>
      </div>
    </section>

    <section class="docs-content section-padding">
      <div class="container-padding">
        <div class="doc-wrapper card-enhanced">
          <div v-if="loading" class="loading">Loading documentation...</div>
          <div v-else-if="error" class="error">{{ error }}</div>
          <article v-else class="markdown-body" v-html="html"></article>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const html = ref('')
const loading = ref(true)
const error = ref('')
const title = ref('Documentation')

async function renderMarkdown(mdText: string) {
  try {
    // Dynamically import a markdown renderer from CDN to avoid adding local deps
    const { marked } = await import('https://cdn.jsdelivr.net/npm/marked@12.0.2/lib/marked.esm.js')
    // Basic sanitization: strip script tags; docs are controlled content in /public
    const safe = mdText.replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
    html.value = marked.parse(safe)
  } catch (e) {
    // Fallback: show as preformatted text
    html.value = `<pre>${mdText.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>`
  }
}

async function loadDoc() {
  loading.value = true
  error.value = ''
  html.value = ''

  const page = (route.params.page as string) || 'index'
  title.value = pageToTitle(page)
  // Update browser tab title
  try {
    document.title = `${title.value} - Mycelium Cloud`
  } catch (_) {}
  const url = `/docs/${page}.md`
  try {
    const res = await fetch(url, { cache: 'no-cache' })
    if (!res.ok) throw new Error(`Not found: ${url}`)
    const md = await res.text()
    await renderMarkdown(md)
  } catch (e: any) {
    error.value = e?.message || 'Failed to load documentation.'
  } finally {
    loading.value = false
  }
}

function pageToTitle(p: string): string {
  const map: Record<string, string> = {
    'index': 'Documentation',
    'getting-started': 'Getting Started',
    'tutorial': 'Platform Tutorial',
    'api-reference': 'API Reference',
    'faq': 'FAQ',
    'architecture': 'Architecture',
    'best-practices': 'Best Practices'
  }
  return map[p] || p.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

onMounted(loadDoc)
watch(() => route.fullPath, loadDoc)
</script>

<style scoped>
.docs-page {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
}
.hero-section { min-height: auto; display:block; padding: 5rem 0 1rem; }
.hero-content { max-width: 1200px; margin: 0 auto; padding: 0 2rem; text-align: left; }
.hero-title { font-size: clamp(1.6rem, 3vw, 2.2rem); line-height: 1.2; color:#fff; font-weight:500; margin: 0 0 0.25rem; word-break: break-word; hyphens: auto; }
.section-subtitle { color:#60a5fa; opacity:0.9; margin: 0; }
.back-actions {
  width: 100%;
  display: flex;
  justify-content: flex-start;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem; /* align with docs-content margins */
  margin-top: 0.75rem;
}
.back-btn { margin-top: 0.5rem; }
.docs-content { padding: 2rem 0 4rem; margin: 0 2rem; }
.doc-wrapper { padding: 1.5rem; }
.loading, .error { color:#fff; }

/* Basic GitHub-like markdown styling */
.markdown-body { color: #e5e7eb; line-height: 1.7; }
.markdown-body h1:first-of-type { display: none; }
.markdown-body h1, .markdown-body h2, .markdown-body h3 { color:#fff; margin-top: 1.6rem; }
.markdown-body pre { background: #0b1220; padding: 1rem; border-radius: 8px; overflow:auto; }
.markdown-body code { background: rgba(96,165,250,0.15); padding: 0.2rem 0.4rem; border-radius: 4px; }
.markdown-body a { color: #93c5fd; }
.markdown-body table { width: 100%; border-collapse: collapse; }
.markdown-body th, .markdown-body td { border: 1px solid #334155; padding: 0.5rem; }
</style>
