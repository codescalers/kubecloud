export interface DashboardLayoutCtx {
  drawer: DrawerCtx
  container: ContainerCtx
}

export interface DrawerCtx {
  isOpen: Ref<boolean>
  open(): void
  close(): void
}

export interface ContainerCtx {
  isFluid: Ref<boolean>
  fluidize(): void
  containerize(): void
}

export const DashboardLayoutCtxKey: InjectionKey<DashboardLayoutCtx> = Symbol("DashboardLayoutCtx")
