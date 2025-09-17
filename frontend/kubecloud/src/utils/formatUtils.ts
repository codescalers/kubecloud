/**
 * Format statistics data for display cards
 * @param stats - Raw statistics object
 * @returns Formatted statistics for display
 */
export interface FormattedStat {
  label: string
  value: string
}

export function formatStatsForCards(stats: {
  ssd: number
  up_nodes: number
  countries: number
  cores: number
  total_users?: number
  total_clusters?: number
}): FormattedStat[] {
  return [
    {
      label: 'SSD Storage',
      value: Math.round(stats.ssd).toLocaleString() + ' GB',
    },
    {
      label: 'Active Nodes',
      value: stats.up_nodes.toString(),
    },
    {
      label: 'Countries',
      value: stats.countries.toString(),
    },
    {
      label: 'CPU Cores',
      value: stats.cores.toString(),
    }
  ]
}

