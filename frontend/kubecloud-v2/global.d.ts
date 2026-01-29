import "axios"

declare module "axios" {
  interface AxiosRequestConfig {
    unauthenticated?: boolean
    notify?: boolean
  }
}

declare global {
  interface Window {
    THREE?: typeof import("three")
    VANTA?: {
      DOTS?: (options: Record<string, any>) => void
      GLOBE?: (options: Record<string, any>) => void
      BIRDS?: (options: Record<string, any>) => void
      HALO?: (options: Record<string, any>) => void
    }
  }
}
