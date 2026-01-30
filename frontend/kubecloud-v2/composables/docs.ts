import yaml from "js-yaml"
import { marked } from "marked"

export interface Doc {
  title: string
  icon: string
  file: string
  path: string
  content: string
  md: {
    html: string
    tableOfContent: []
  }
}

export const useDocs = createGlobalState(() => {
  return useAsyncState(async () => {
    const res = await fetch("/docs/docs.yaml")
    const data = await res.text()
    const docs = yaml.load(data) as Doc[]
    const contents = await Promise.all(
      docs
        .map(doc => fetch(`/docs/${doc.file}`)
          .then(res => res.text())),
    )

    marked.use({
      renderer: {
        blockquote({ tokens }) {
          const text = this.parser.parse(tokens)
          return `<blockquote class="d-block px-4 py-6 border-s-lg border-primary bg-surface mb-4 overflow-auto" style="--v-border-opacity: 1;">${text}</blockquote>`
        },
      },
    })
    return docs.map((doc, index) => {
      const content = contents[index] ?? ""
      return {
        ...doc,
        content,
        md: {
          html: marked.parse(content),
          tableOfContent: [],
        },
      }
    })
  }, [])
})
