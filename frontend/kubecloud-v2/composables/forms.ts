export const useRegisterationFormData = createGlobalState(() => {
  return useSessionStorage<{
    username: string
    email: string
    password: string
  } | null>("registrationFormData", null, {
    serializer: {
      read: JSON.parse,
      write: JSON.stringify,
    },
  })
})
