<template>
  <div class="mt-16">
    <v-container fluid class="pb-12">
      <div class="text-center pa-8 mt-16">
        <h1 class="text-h3 font-weight-bold mb-2">Documentation</h1>
        <p class="text-h6 text-medium-emphasis">Find guides, and tutorials to get the most out of Mycelium Cloud.</p>
      </div>

      <v-row>
        <!-- Sidebar -->
        <v-col cols="12" md="3" lg="2">
          <v-card class="elevation-0 rounded-xl border pa-6 sidebar-card" flat style="position: sticky; top: 80px; max-height: calc(100vh - 80px); overflow: auto; background: transparent;">
            <!-- Documentation Navigation -->
            <div class="mb-6">
              <h3 class="text-h6 font-weight-bold mb-4 text-primary">Documentation</h3>
              <v-list nav density="comfortable" class="pa-0">
                <v-list-item
                  v-for="doc in docs"
                  :key="doc.slug"
                  :active="currentDoc === doc.slug"
                  @click="loadDocument(doc.slug)"
                  class="rounded-lg mb-2"
                >
                  <template v-slot:prepend>
                    <v-icon :icon="doc.icon" size="20" color="primary" class="mr-3"></v-icon>
                  </template>
                  <v-list-item-title class="text-body-1 font-weight-medium">{{ doc.title }}</v-list-item-title>
                </v-list-item>
              </v-list>
            </div>

            <!-- Table of Contents -->
            <div v-if="toc.length > 0" class="toc-section">
              <v-divider class="my-6"></v-divider>
              <div class="mb-4">
                <h4 class="text-subtitle-2 font-weight-bold text-uppercase text-primary">Table of Contents</h4>
              </div>
              <v-list density="comfortable" class="pa-0">
                <v-list-item
                  v-for="item in toc"
                  :key="item.id"
                  @click="scrollToHeading(item.id)"
                  class="py-2 rounded toc-item"
                >
                  <v-list-item-title class="text-body-2">{{ item.text }}</v-list-item-title>
                </v-list-item>
              </v-list>
            </div>
          </v-card>
        </v-col>

        <!-- Main Content -->
        <v-col cols="12" md="9" lg="10" class="pa-0">
          <v-container fluid>
            <!-- Loading State -->
            <v-card v-if="loading" class="pa-8 text-center mb-6" variant="tonal">
              <v-progress-circular indeterminate color="primary" size="64" class="mb-4"></v-progress-circular>
              <v-card-text class="text-h6">Loading documentation...</v-card-text>
            </v-card>

            <!-- Error State -->
            <v-card v-else-if="error" class="pa-8 text-center mb-6" variant="tonal" color="error">
              <v-icon icon="mdi-alert-circle" size="48" class="mb-4"></v-icon>
              <v-card-title class="mb-2">Error Loading Documentation</v-card-title>
              <v-card-text class="mb-4">{{ error }}</v-card-text>
              <v-btn @click="loadDocument(currentDoc)" color="primary">Retry</v-btn>
            </v-card>

            <!-- Content -->
            <div v-else-if="renderedMarkdown">
              <!-- Markdown Content -->
              <v-card class="pa-8 mb-8 rounded-xl" variant="tonal">
                <div class="markdown-content" v-html="renderedMarkdown"></div>
              </v-card>
            </div>
          </v-container>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, watch } from 'vue'
import { marked } from 'marked'

const loading = ref(false)
const error = ref('')
const renderedMarkdown = ref('')
const currentDoc = ref('')
const toc = ref<Array<{ id: string; text: string }>>([])

const docs = [
  {
    icon: 'mdi-rocket-launch',
    title: 'Getting Started',
    slug: 'getting-started',
    path: '/md/getting-started.md'
  },
  {
    icon: 'mdi-book-open-page-variant',
    title: 'Tutorials',
    slug: 'tutorial',
    path: '/md/tutorial.md'
  }
]

// Configure marked with syntax highlighting
marked.setOptions({
  breaks: true,
  gfm: true
})

const loadDocument = async (slug: string) => {
  const doc = docs.find(d => d.slug === slug)
  if (!doc) return

  loading.value = true
  error.value = ''
  currentDoc.value = slug

  try {
    const response = await fetch(doc.path)
    if (!response.ok) throw new Error(`Failed to load ${doc.title}`)

    const md = await response.text()
    renderedMarkdown.value = await marked.parse(md)

  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error'
  } finally {
    loading.value = false
  }
}

const generateTOC = () => {
  const markdownElement = document.querySelector('.markdown-content')
  if (!markdownElement) return
  
  const headings = markdownElement.querySelectorAll('h2')
  toc.value = Array.from(headings).map((heading, index) => {
    const id = heading.textContent?.toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^\w-]/g, '')
      .replace(/^-+|-+$/g, '') || `heading-${index}`
    heading.id = id
    return { id, text: heading.textContent || '' }
  })
}

const scrollToHeading = (id: string) => {
  const element = document.getElementById(id)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

// Handle copy buttons and navigation links
const setupCopyButtons = () => {
  document.addEventListener('click', async (e) => {
    const target = e.target as HTMLElement
    
    // Handle copy button clicks
    const btn = target.closest('.copy-btn')
    if (btn) {
      const pre = btn.closest('pre')
      const text = pre?.textContent || ''
      
      try {
        await navigator.clipboard.writeText(text)
        btn.innerHTML = '<span class="mdi mdi-check"></span>'
        setTimeout(() => btn.innerHTML = '<span class="mdi mdi-content-copy"></span>', 1500)
      } catch (err) {
        console.error('Copy failed:', err)
      }
      return
    }
    
    // Handle navigation links to sidebar items
    const link = target.closest('a[href^="#"]') as HTMLAnchorElement | null
    if (link && link.href.endsWith('#getting-started')) {
      e.preventDefault()
      loadDocument('getting-started')
    }
  })
}

// Add copy buttons to code blocks
const addCopyButtons = () => {
  document.querySelectorAll('.markdown-content pre:not(.has-copy)').forEach(pre => {
    pre.classList.add('has-copy')
    pre.innerHTML += '<button class="copy-btn"><span class="mdi mdi-content-copy"></span></button>'
  })
}


// Initialize
onMounted(() => {
  if (docs.length > 0) {
    loadDocument(docs[0].slug)
  }
  setupCopyButtons()
})

// Update when content changes
watch(renderedMarkdown, async () => {
  await nextTick()
  generateTOC()
  addCopyButtons()
})
</script>

<style>
.markdown-content {
  font-size: 1.125rem !important; /* 18px */
  line-height: 1.8 !important;
  color: rgb(var(--v-theme-on-surface)) !important;
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3,
.markdown-content h4,
.markdown-content h5,
.markdown-content h6 {
  margin-top: 2rem;
  margin-bottom: 1rem;
  scroll-margin-top: 96px;
  color: rgb(var(--v-theme-on-surface));
}

.markdown-content h1 { 
  font-size: 1.875rem; 
  font-weight: 700; 
  margin-top: 0;
}

.markdown-content h2 { 
  font-size: 1.5rem; 
  font-weight: 700; 
}

.markdown-content p { 
  margin: 0.75rem 0 1.5rem 0; 
  font-size: 1.125rem;
}

.markdown-content ul,
.markdown-content ol { 
  margin: 0.75rem 0 1.5rem 1.5rem; 
  font-size: 1.125rem;
}

.markdown-content li {
  margin: 0.5rem 0;
}

.markdown-content table { 
  margin: 1.5rem 0; 
  font-size: 1rem;
}

.markdown-content blockquote {
  margin: 1.5rem 0;
  padding: 1rem 1.5rem;
  border-left: 4px solid rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.05);
  font-size: 1.125rem;
}
</style>

<style scoped>
:deep(.sidebar-card .v-list),
:deep(.sidebar-card .v-list-item) {
  background: transparent;
}

:deep(.sidebar-card .v-divider) {
  opacity: 0.16;
}

/* Code blocks */
:deep(.markdown-content pre) {
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1.5rem 0;
  overflow-x: auto;
  position: relative;
  font-family: monospace;
  font-size: 0.9rem;
  color: #d4d4d4;
}

:deep(.markdown-content code:not(pre code)) {
  background: rgba(var(--v-theme-secondary), 0.15);
  border: 1px solid rgba(var(--v-theme-secondary), 0.3);
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.85em;
  color: rgb(var(--v-theme-secondary));
  font-weight: 500;
}

:deep(.markdown-content .copy-btn) {
  position: absolute;
  top: 12px;
  right: 12px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  padding: 8px;
  cursor: pointer;
  color: #d4d4d4;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>