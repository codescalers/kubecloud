import { v4 } from "uuid"

export interface ClusterForm {
  name: string
  masters: ClusterNode[]
  workers: ClusterNode[]
}

export interface ClusterNode {
  id: string
  name: string
  permanent: boolean
  useFullNodeCapabilities: boolean
  cpu: number
  memory: number
  disk: number
  nodeId: number | null
  sshKeys: number[]
}

export function createClusterNode(opts: Partial<ClusterNode> = {}): ClusterNode {
  return {
    id: v4(),
    name: `engine${Math.floor(Math.random() * 1000)}`,
    permanent: false,
    useFullNodeCapabilities: true,
    cpu: 2,
    memory: 4,
    disk: 25,
    nodeId: null,
    sshKeys: [0],
    ...opts,
  }
}

export function isValidClusterNode(node: ClusterNode): boolean {
  return node.name.length >= 3
    && (node.useFullNodeCapabilities || (node.cpu > 0 && node.memory > 0 && node.disk > 0))
    && node.sshKeys.length > 0
}
