export const useTokens = createGlobalState(() => ({
  accessToken: useLocalStorage<string>("accessToken", "", { writeDefaults: false }),
  refreshToken: useLocalStorage<string>("refreshToken", "", { writeDefaults: false }),
}))
