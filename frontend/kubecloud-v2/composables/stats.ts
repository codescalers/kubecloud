export interface StatsResource {
  id: string
  title: string
  icon: string
  color: string
  value: string | number
}

export type Resources =
  | "users"
  | "clusters"
  | "nodes"
  | "countries"
  | "balance"
  | "memory"
  | "storage"

export interface UseStatsOptions {
  exclude?: Resources[]
}

export const useStats = (options?: UseStatsOptions) => {
  const api = useApi()

  const { state, isLoading } = useAsyncState(async () => {
    const { data } = await api.admin.getStats({ unauthenticated: true })
    return data.data
  }, null)

  const stats = computed(() => {
    const s = state.value
    if (!s) {
      return []
    }

    const resources: StatsResource[] = []
    const exclude = options?.exclude ?? []

    if (!exclude.includes("users")) {
      resources.push({
        id: "users",
        title: "Total Users",
        icon: "mdi-account-group-outline",
        color: "#359EFF",
        value: s.total_users ?? 0,
      })
    }

    if (!exclude.includes("clusters")) {
      resources.push({
        id: "clusters",
        title: "Active Clusters",
        icon: "mdi-server-outline",
        color: "#607AFB",
        value: s.total_clusters ?? 0,
      })
    }

    if (!exclude.includes("nodes")) {
      resources.push({
        id: "nodes",
        title: "Up Nodes",
        icon: "mdi-server-network-outline",
        color: "#39E079",
        value: s.up_nodes ?? 0,
      })
    }

    if (!exclude.includes("countries")) {
      resources.push({
        id: "countries",
        title: "Countries",
        icon: "mdi-earth",
        color: "#D0BB95",
        value: s.countries ?? 0,
      })
    }

    if (!exclude.includes("balance")) {
      resources.push({
        id: "balance",
        title: "System Balance",
        icon: "mdi-wallet-bifold",
        color: "#4c16c9",
        value: `$${Math.floor(s.system_account_balance ?? 0)}`,
      })
    }

    if (!exclude.includes("memory")) {
      resources.push({
        id: "memory",
        title: "Total Memory",
        icon: "mdi-memory",
        color: "#EA2831",
        value: s.cores ?? 0,
      })
    }

    if (!exclude.includes("storage")) {
      resources.push({
        id: "storage",
        title: "Total Storage",
        icon: "mdi-harddisk",
        color: "#FAC638",
        value: `${(Math.floor((s.ssd ?? 0) / 1024) * 100) / 100} TB`,
      })
    }

    return resources
  })

  return { isLoading, stats }
}
