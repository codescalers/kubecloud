export default defineNuxtRouteMiddleware(() => {
  const { accessToken } = useTokens()
  if (accessToken.value) {
    return navigateTo("/dashboard")
  }
})
