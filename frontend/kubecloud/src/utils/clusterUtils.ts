import { VALIDATION_RULES, validateField } from './validation';
import { useClusterStore } from '../stores/clusters';

/**
 * Generate a random cluster name using adjectives and nouns
 * @returns Random cluster name
 */
export function generateClusterName(): string {

  const nouns = [
    'cluster', 'cloud', 'node', 'server', 'engine', 'core', 'hub', 'nexus',
    'forge', 'vault', 'tower', 'citadel', 'fortress', 'sanctuary', 'haven',
    'realm', 'domain', 'sphere', 'matrix', 'grid', 'system'
  ]

  const randomNoun = nouns[Math.floor(Math.random() * nouns.length)]
  const randomNumber = Math.floor(Math.random() * 999) + 1

  return `${randomNoun}${randomNumber}`
}

/**
 * Validate cluster name according to backend requirements and check uniqueness
 * @param name - Cluster name to validate
 * @returns Object with isValid boolean and error message
 */
export function validateClusterName(name: string): { isValid: boolean; error: string } {
  const result = validateField({
    value: name,
    rules: VALIDATION_RULES.CLUSTER_NAME,
    fieldName: 'Cluster name'
  });

  if (!result.isValid) {
    return {
      isValid: false,
      error: result.errors.length > 0 ? result.errors[0] : ''
    };
  }

  // Check for uniqueness in existing clusters
  const clusterStore = useClusterStore();
  const existingNames = clusterStore.clusters.map(c => c.cluster.name);
  if (existingNames.includes(name)) {
    return {
      isValid: false,
      error: 'Cluster name already exists'
    };
  }

  return {
    isValid: true,
    error: ''
  };
}

/**
 * Get node information string for display
 * @param nodeId - Node ID
 * @param availableNodes - Array of available nodes
 * @returns Formatted node info string
 */
export function getNodeInfo(nodeId: number | null, availableNodes: any[]): string {
  if (nodeId == null) return ''
  const node = availableNodes.find(n => n.nodeId === nodeId)
  if (!node) return ''
  return `${node.cpu} vCPU, ${node.ram}GB RAM${node.gpu ? ', GPU Available' : ''}`
}

/**
 * Get SSH key name by ID
 * @param keyId - SSH key ID
 * @param availableSshKeys - Array of available SSH keys
 * @returns SSH key name or 'Unknown'
 */
export function getSshKeyName(keyId: number, availableSshKeys: any[]): string {
  const key = availableSshKeys.find(k => k.ID === keyId)
  return key ? key.name : 'Unknown'
}
