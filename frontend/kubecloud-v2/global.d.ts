import "axios"

declare module "axios" {
  interface AxiosRequestConfig {
    unauthenticated?: boolean
    notify?: boolean
  }
}
