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
