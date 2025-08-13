/**
 * Generate a random cluster name using short words and numbers
 * @returns Random cluster name (max 8 characters to allow for node suffixes)
 */
export function generateClusterName(): string {
  const nouns = [
    'app', 'web', 'api', 'db', 'dev', 'prod', 'test', 'demo',
    'core', 'hub', 'net', 'sys', 'ops', 'run', 'box', 'lab'
  ]

  const randomNoun = nouns[Math.floor(Math.random() * nouns.length)]
  const randomNumber = Math.floor(Math.random() * 99) + 1

  return `${randomNoun}${randomNumber}`
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
