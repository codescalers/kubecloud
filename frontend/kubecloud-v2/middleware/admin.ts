export default defineNuxtRouteMiddleware(() => {
  const { accessToken } = useTokens()
  if (!accessToken.value) {
    return navigateTo("/sign-in")
  }

  console.warn("should be admin")

  // TODO: Check if user is admin
})
