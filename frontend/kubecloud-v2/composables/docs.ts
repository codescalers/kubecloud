import fm from "front-matter"
import hljs from "highlight.js"
import yaml from "js-yaml"
import { Marked } from "marked"
import { markedHighlight } from "marked-highlight"
import "highlight.js/styles/atom-one-dark.css"

export interface Doc {
  title: string
  icon: string
  file: string
  path: string
  content: string
  md: {
    html: string
    tableOfContent: { id: string, content: string }[]
    attributes: Record<string, string>
  }
}

export const marked = new Marked(
  markedHighlight({
    emptyLangClass: "hljs",
    langPrefix: "hljs language-",
    highlight(code, lang) {
      const language = hljs.getLanguage(lang) ? lang : "plaintext"
      return hljs.highlight(code, { language }).value
    },
  }),
)

export const renderer = new marked.Renderer()

renderer.code = function ({ text }) {
  const parts = text.split("\n")
  const count = parts.length
  const size = 29 + count.toString().length * 15
  const lines = parts.map((_, i) => `<span class="text-accent">${i + 1}</span>`)

  return `
    <pre 
      class="my-6 border rounded d-flex"
      style="--v-border-color: 255, 255, 255; --v-border-opacity: 0.12"
      >
      <code class="py-4 d-inline-block border-e bg-surface text-center" style="width: ${size}px;">${lines.join("\n")}</code>
      <code class="hljs bg-surface" style="width: calc(100% - ${size}px);">${parts.join("\n")}</code>
    </pre>
  `
}

renderer.blockquote = function ({ tokens }) {
  const content = this.parser.parse(tokens)
  return `<blockquote class="my-6 px-6 pa-4 border-s-lg border-primary rounded" style="--v-border-opacity: 1; background: rgba(var(--v-theme-primary), 0.12);">${content}</blockquote>`
}

function getIdFromTitle(title: string) {
  let id = ""
  for (const char of title.toLowerCase()) {
    const code = char.charCodeAt(0)
    const isAlpha = code >= 97 && code <= 122
    const isAllowedSpace = code === 32 && id.length > 0 && id[id.length - 1] !== " "
    if (isAlpha || isAllowedSpace) {
      id += char
    }
  }

  return id.replaceAll(" ", "-")
}

let tableOfContent: Doc["md"]["tableOfContent"] = []

const linkIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="currentColor" d="M10.59 13.41c.41.39.41 1.03 0 1.42c-.39.39-1.03.39-1.42 0a5.003 5.003 0 0 1 0-7.07l3.54-3.54a5.003 5.003 0 0 1 7.07 0a5.003 5.003 0 0 1 0 7.07l-1.49 1.49c.01-.82-.12-1.64-.4-2.42l.47-.48a2.98 2.98 0 0 0 0-4.24a2.98 2.98 0 0 0-4.24 0l-3.53 3.53a2.98 2.98 0 0 0 0 4.24m2.82-4.24c.39-.39 1.03-.39 1.42 0a5.003 5.003 0 0 1 0 7.07l-3.54 3.54a5.003 5.003 0 0 1-7.07 0a5.003 5.003 0 0 1 0-7.07l1.49-1.49c-.01.82.12 1.64.4 2.43l-.47.47a2.98 2.98 0 0 0 0 4.24a2.98 2.98 0 0 0 4.24 0l3.53-3.53a2.98 2.98 0 0 0 0-4.24a.973.973 0 0 1 0-1.42"/></svg>`
renderer.heading = function ({ depth, tokens }) {
  const content = this.parser.parseInline(tokens)
  let id = ""
  let a = ""
  if (depth === 2) {
    id = getIdFromTitle(content)
    // .toLowerCase().replaceAll(":", "").replaceAll(" ", "-")
    a = `<a href="#${id}" class="d-inline-block mr-2 text-primary opacity-50">${linkIcon}</a>`
    tableOfContent.push({ id, content })
  }

  const lv = Math.min(depth + 3, 6)
  return `
        <h${depth} id="${id}" class="text-white text-h${lv} mb-4">
          ${a}${content}
        </h${depth}>
      `
}

renderer.paragraph = function ({ tokens }) {
  const content = this.parser.parseInline(tokens)
  return `<p class="text-body-1 text-accent mt-3 mb-6">${content}</p>`
}

renderer.listitem = function (x) {
  const content = this.parser.parse(x.tokens)
  return `
        <li class="mb-2 text-white">
          <span class="text-body-1 text-accent">${content}</span>
        </li>
      `
}

renderer.list = function ({ ordered, items }) {
  const tag = ordered ? "ol" : "ul"
  const body = items.map(item => this.listitem(item)).join("")
  return `<${tag} class="mt-4 mb-6 pl-4" style="list-style-type: square">${body}</${tag}>`
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

renderer.strong = function ({ tokens }) {
  const content = this.parser.parseInline(tokens)
  return `<strong class="text-white text-body-1">${content}</strong>`
}

renderer.codespan = function ({ text }) {
  return `<code
          class="text-primary text-body-1 border border-primary py-1 px-2 rounded"
          style="--v-border-opacity: 0.12; background-color: rgba(var(--v-theme-primary), var(--v-border-opacity))"
        >${text}</code>`
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

    // let tableOfContent: Doc["md"]["tableOfContent"] = []

    return docs.map((doc, index) => {
      tableOfContent = []
      const { attributes, body } = fm(contents[index] ?? "") as { attributes: Record<string, string>, body: string }
      const content = body
      return {
        ...doc,
        content,
        md: {
          html: marked.parse(content, { renderer }),
          tableOfContent,
          attributes,
        },
      }
    })
  }, [])
})
