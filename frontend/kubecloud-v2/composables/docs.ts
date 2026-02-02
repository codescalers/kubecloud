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
  // const router = useRouter()

  return useAsyncState(async () => {
    const res = await fetch("/docs/docs.yaml")
    const data = await res.text()
    const docs = yaml.load(data) as Doc[]
    const contents = await Promise.all(
      docs
        .map(doc => fetch(`/docs/${doc.file}`)
          .then(res => res.text())),
    )

    const renderer = new marked.Renderer()

    renderer.heading = function ({ depth, tokens }) {
      const content = this.parser.parseInline(tokens)
      const lv = Math.min(depth + 3, 6)
      return `<h${lv} class="text-h${lv} mb-2">${content}</h${lv}>`
    }

    renderer.paragraph = function ({ tokens }) {
      const content = this.parser.parseInline(tokens)
      return `<p class="text-body-1 text-accent mt-2 mb-4">${content}</p>`
    }

    renderer.listitem = function (x) {
      // console.log(x.tokens)

      // return "item"
      const content = this.parser.parse(x.tokens)
      return `<li class="text-body-1 text-accent">${content}</li>`
    }

    renderer.list = function ({ ordered, items }) {
      const tag = ordered ? "ol" : "ul"
      const body = items.map(item => this.listitem(item)).join("")
      return `<${tag} class="mt-2 mb-4 pl-4" style="list-style-type: square">${body}</${tag}>`
    }

    renderer.link = function ({ href, tokens }) {
      const content = this.parser.parseInline(tokens)

      // if (href.startsWith("/")) {
      // const a = document.createElement("a")
      // a.href = href
      // a.textContent = content
      // a.addEventListener("click", (e) => {
      //   e.preventDefault()
      //   console.log(e)
      // })
      // return a.outerHTML
      // }

      ;(window as any).xonClick = function (e: Event, _: HTMLAnchorElement) {
        e.preventDefault()
        // console.log({ event, target })
      }

      return `<a class="text-link" href="${href}" onclick="xonClick(event, this);">${content}</a>`
    }

    renderer.codespan = function ({ text }) {
      return `<code
          class="text-primary text-body-1 border border-primary py-1 px-2 rounded"
          style="background-color: rgba(var(--v-theme-primary), var(--v-border-opacity))"
        >${text}</code>`
    }

    /* Headings */
    // renderer.heading = ({ text, depth: level }) => {
    //   return `<h${level} class="md-heading text-h${Math.min(level + 3, 6)}">${text}</h${level}>`
    // }

    /* Paragraphs */
    // renderer.paragraph = (text) => {
    //   return `<p class="md-paragraph">${text}</p>`
    // }

    /* Links */
    // renderer.link = ({ href, title, text }) => {
    //   const t = title ? ` title="${title}"` : ""
    //   return `<a class="md-link" href="${href}"${t} target="_blank" rel="noopener noreferrer">${text}</a>`
    // }

    /* Lists */
    // renderer.list = ({ ordered, items }) => {
    //   const tag = ordered ? "ol" : "ul"
    //   const body = items.map(item => `<li class="md-list-item">${item}</li>`).join("")
    //   return `<${tag} class="md-list">${body}</${tag}>`
    // }

    // renderer.listitem = (text) => {
    //   return `<li class="md-list-item">${text}</li>`
    // }

    /* Blockquotes */
    // renderer.blockquote = (quote) => {
    //   return `<blockquote class="md-blockquote">${quote}</blockquote>`
    // }

    /* Inline code */
    // renderer.codespan = (code) => {
    //   return `<code class="md-inline-code">${code}</code>`
    // }

    /* Code blocks */
    //   renderer.code = ({ text, lang }) => {
    //     const langClass = lang ? ` lang-${lang}` : ""
    //     return `
    //   <pre class="md-code-block${langClass}">
    //     <code class="md-code">${text}</code>
    //   </pre>
    // `
    //   }

    /* Tables */
    //   renderer.table = ({ header, rows }) => {
    //     const body = marked.parseInline(rows)
    //     return `
    //   <table class="md-table">
    //     <thead class="md-table-head">${header}</thead>
    //     <tbody class="md-table-body">${body}</tbody>
    //   </table>
    // `
    //   }

    // renderer.tablerow = (content) => {
    //   return `<tr class="md-table-row">${content}</tr>`
    // }

    // renderer.tablecell = (content, flags) => {
    //   const tag = flags.header ? "th" : "td"
    //   return `<${tag} class="md-table-cell">${content}</${tag}>`
    // }

    /* Images */
    // renderer.image = ({ href, title, text }) => {
    //   const t = title ? ` title="${title}"` : ""
    //   return `<img class="md-image" src="${href}" alt="${text}"${t} />`
    // }

    /* Horizontal rule */
    // renderer.hr = () => {
    //   return `<hr class="md-hr" />`
    // }

    /* Strong / emphasis */
    // renderer.strong = (text) => {
    //   return `<strong class="md-strong">${text}</strong>`
    // }

    // renderer.em = (text) => {
    //   return `<em class="md-em">${text}</em>`
    // }

    // function e(name: string, attrs: Record<string, string>) {
    //   const content = attrs.content ?? ""
    //   delete attrs.content
    //   return `<${name} ${Object.entries(attrs).map(([key, value]) => `${key}="${value}"`).join(" ")}>${content}</${name}>`
    // }

    // marked.use({
    //   renderer: {
    //     blockquote({ tokens }) {
    //       const text = this.parser.parse(tokens)
    //       return `
    //         <blockquote
    //           class="pa-4 border-s-lg border-primary bg-surface"
    //           style="--v-border-opacity: 1"
    //         >
    //           ${text}
    //         </blockquote>
    //       `
    //     },
    //     code(code) {
    //       console.log(code)

    //       if (!code.type) {
    //         console.log({ code })
    //       }
    //       return `
    //         <pre class="bg-red overflow-y-auto"><code class="language-${code.lang}">${code.text}</code></pre>
    //       `
    //     },
    //   },
    // })
    return docs.map((doc, index) => {
      const content = contents[index] ?? ""
      return {
        ...doc,
        content,
        md: {
          html: marked.parse(content, { renderer }),
          tableOfContent: [],
        },
      }
    })
  }, [])
})
