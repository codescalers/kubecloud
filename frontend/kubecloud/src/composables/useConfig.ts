export function useConfig() {
  function isBalanceCheckDisabled(): boolean {
    // Disable balance check when running on development network
    return import.meta.env.VITE_NETWORK === 'dev'
  }

  return {
    isBalanceCheckDisabled
  }
}
