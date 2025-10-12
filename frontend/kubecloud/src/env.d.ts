/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_STRIPE_PUBLISHABLE_KEY: string
  readonly VITE_BASE_URL: string
  readonly VITE_NETWORK: string
  readonly VITE_MIN_BALANCE: number
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
