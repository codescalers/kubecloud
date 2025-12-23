export default defineNuxtRouteMiddleware((to) => {
  if (to.meta.middleware !== undefined) {
    return
  }

  const { accessToken } = useTokens()
  if (!accessToken.value) {
    return navigateTo("/login")
  }
})
