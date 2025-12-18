import "axios"

declare module "axios" {
  interface AxiosRequestConfig {
    _flags?: {
      unauthenticated?: boolean
    }
  }
}
