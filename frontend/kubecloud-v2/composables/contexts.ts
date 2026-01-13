import type { HandlersCreditRequestInput, HandlersNodeInput, ServicesClusterData, ServicesUserWithUSDBalance } from "../generated/api"

// Dashboard Layout Context
export interface DashboardLayoutCtx {
  drawer: DrawerCtx
  container: ContainerCtx
}

export interface DrawerCtx {
  isOpen: Ref<boolean>
  open: () => void
  close: () => void
}

export interface ContainerCtx {
  isFluid: Ref<boolean>
  fluidize: () => void
  containerize: () => void
}

export const DashboardLayoutCtxKey: InjectionKey<DashboardLayoutCtx> = Symbol("DashboardLayoutCtx")

// User Dialog Context

export interface UserDialogCtx {
  credit: (user: ServicesUserWithUSDBalance) => Promise<HandlersCreditRequestInput | undefined>
  drain: (user: ServicesUserWithUSDBalance) => Promise<boolean>
  remove: (user: ServicesUserWithUSDBalance) => Promise<boolean>
}

export const UserDialogCtxKey: InjectionKey<UserDialogCtx> = Symbol("UserDialogCtx")

// Deployment Dialog Context
export interface DeploymentDialogCtx {
  addNode: (deployment: ServicesClusterData) => Promise<HandlersNodeInput | undefined>
  delete: (deployment: ServicesClusterData) => Promise<boolean>
}

export const DeploymentDialogCtxKey: InjectionKey<DeploymentDialogCtx> = Symbol("DeploymentDialogCtx")
