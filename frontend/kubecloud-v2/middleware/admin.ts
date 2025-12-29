export default defineNuxtRouteMiddleware(() => {
  const { accessToken } = useTokens()
  if (!accessToken.value) {
    return navigateTo("/sign-in")
  }

  console.log("should be admin")

  // TODO: Check if user is admin
})
