/// <reference types="vite/client" />

declare global {
  const __ENV__: {
    VITE_API_BASE_URL: string
    VITE_NETWORK: string
    VITE_STRIPE_PUBLISHABLE_KEY: string
  }
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
  readonly VITE_NETWORK: string
  readonly VITE_STRIPE_PUBLISHABLE_KEY: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
