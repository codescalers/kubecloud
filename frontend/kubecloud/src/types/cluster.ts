// Types for the new backend Cluster and Node payload structure

export type NodeType = 'master' | 'worker' | 'leader';

export interface ClusterNode {
  name: string;
  type: NodeType;
  node_id: number;
  cpu: number;
  memory: number;      // MB
  root_size: number;    // MB
  data_disks: number[];    // MB
  env_vars: Record<string, string>;
  ip?: string;
  mycelium_ip?: string;
  planetary_ip?: string;
  contract_id?: number;
}

export interface Cluster {
  name: string;
  token: string;
  nodes: ClusterNode[];
} 