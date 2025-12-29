export const useStatusColor = (status: () => string) => {
  return computed(() => {
    switch (status()) {
      case "running":
        return "primary"
      case "completed":
        return "success"
      case "failed":
        return "error"
      default:
        return "warning"
    }
  })
}
