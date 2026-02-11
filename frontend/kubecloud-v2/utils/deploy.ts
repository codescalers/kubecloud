import type { HandlersNodesWithDiscount } from "~/generated/api"
import { v4 } from "uuid"

export interface ClusterForm {
  name: string
  token: string
  region: string | null
  masters: ClusterNode[]
  workers: ClusterNode[]
}

function _genName(prefix: string) {
  return `${prefix}${Math.floor(Math.random() * 1000)}`
}

export function createClusterForm(): ClusterForm {
  return {
    name: _genName("cluster"),
    token: v4(),
    region: null,
    masters: [createClusterNode({ type: "leader", name: "Leader" })],
    workers: [],
  }
}

export interface ClusterNode {
  id: string
  type: string
  name: string
  useFullNodeCapabilities: boolean
  cpu: number
  memory: number
  disk: number
  node: null | {
    id: number
    raw: HandlersNodesWithDiscount
    valid: boolean
  }
  sshKeys: number[]
}

export function createClusterNode(opts: Partial<ClusterNode> = {}): ClusterNode {
  return {
    id: v4(),
    type: "worker",
    name: _genName(opts.type ?? "worker"),
    useFullNodeCapabilities: true,
    cpu: 2,
    memory: 4,
    disk: 25,
    node: null,
    sshKeys: [0],
    ...opts,
  }
}

export function isValidClusterNode(node: ClusterNode): boolean {
  return node.name.length >= 3
    && (node.useFullNodeCapabilities || (node.cpu > 0 && node.memory > 0 && node.disk > 0))
    && node.sshKeys.length > 0
}
