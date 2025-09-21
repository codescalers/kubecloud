export interface VM {
  id: number
  project_name: string
  vm: {
    name: string
    node_id: number
    cpu: number
    memory: number
    root_size: number
    disk_size: number
    flist?: string
    entrypoint?: string
    env_vars?: Record<string, string>
    status?: string
  }
  created_at: string
}

export interface VMInput {
  name: string
  node_id: number
  cpu: number
  memory: number
  root_size: number
  disk_size: number
  flist?: string
  entrypoint?: string
  env_vars?: Record<string, string>
}
