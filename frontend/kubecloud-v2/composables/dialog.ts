// import type { UseConfirmDialogRevealResult } from "@vueuse/core"

export function useDialog<A = undefined, B = undefined, C = undefined>() {
  const data = ref<A>()
  const { onReveal, ...reset } = useConfirmDialog<A, B, C>()

  onReveal(((d: A) => (data.value = d)) as any)

  return {
    data,
    onReveal,
    ...reset,
  }
}

export function useFilesDialog(options?: Parameters<typeof useFileDialog>[0]) {
  const { onChange, onCancel, open, reset } = useFileDialog(options)
  const files = ref<File[]>([])

  onChange((newFiles) => {
    if (newFiles) {
      const fs = [...files.value]

      for (const newFile of newFiles) {
        const exists = fs.some((f) => {
          return (
            f.name === newFile.name
            && f.size === newFile.size
            && f.lastModified === newFile.lastModified
          )
        })

        if (!exists) {
          fs.push(newFile)
        }
      }

      files.value = fs
    }
  })

  function removeFile(file: File) {
    files.value = files.value.filter(f => f !== file)
  }

  return {
    onChange,
    onCancel,
    files,
    open,
    reset,
    removeFile,
  }
}
