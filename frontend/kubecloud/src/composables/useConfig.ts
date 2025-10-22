export function useConfig() {
  function getMinBalance(): number {
    return Number(import.meta.env.VITE_MIN_BALANCE) || 0
  }

  return {
    getMinBalance
  }
}
