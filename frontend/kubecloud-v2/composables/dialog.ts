export const useDialog = <T>() => {
  const data = ref<T>()
  const { onReveal, ...reset } = useConfirmDialog()

  onReveal((d) => (data.value = d))

  return {
    data,
    onReveal,
    ...reset,
  }
}

export const useFilesDialog = (options?: Parameters<typeof useFileDialog>[0]) => {
  const { onChange, onCancel, open, reset } = useFileDialog(options)
  const files = ref<File[]>([])

  onChange((newFiles) => {
    if (newFiles) {
      const fs = [...files.value]
      outer: for (const newFile of newFiles) {
        for (const file of fs) {
          if (
            file.name === newFile.name &&
            file.size === newFile.size &&
            file.lastModified === newFile.lastModified
          ) {
            continue outer
          }
        }

        fs.push(newFile)
      }

      files.value = fs
    }
  })

  function removeFile(file: File) {
    files.value = files.value.filter((f) => f !== file)
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
