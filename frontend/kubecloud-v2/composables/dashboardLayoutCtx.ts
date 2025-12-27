export interface DashboardLayoutCtx {
  drawer: DrawerCtx
}

export interface DrawerCtx {
  isOpen: Ref<boolean>
  open(): void
  close(): void
}

export const DashboardLayoutCtxKey: InjectionKey<DashboardLayoutCtx> = Symbol("DashboardLayoutCtx")
