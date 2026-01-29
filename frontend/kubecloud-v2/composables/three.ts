export function useVanta(type: "dots" | "globe", el: () => HTMLElement) {
  useHead({
    script: [
      {
        async: true,
        src: `/scripts/vanta.${type}.min.js`,
        defer: true,
        onload() {
          applyVanta(type, el())
        },
      },
    ],
  })

  const script = document.querySelector(`script[src="/scripts/vanta.${type}.min.js"]`)
  if (script) {
    nextTick(() => {
      applyVanta(type, el())
    })
  }
}
