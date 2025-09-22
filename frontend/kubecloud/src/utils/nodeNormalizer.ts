import type { RawNode } from '../types/rawNode';
import type { NormalizedNode } from '../types/normalizedNode';
import type { RentedNode } from '../composables/useNodeManagement';

export function normalizeNode(node: RawNode): NormalizedNode {
  return {
    nodeId: node.nodeId,
    farmId: node.farmId,
    twinId: node.twinId,
    cpu: node.total_resources?.cru ?? 0,
    ram: node.total_resources?.mru ? Math.round(node.total_resources.mru / (1024 * 1024 * 1024)) : 0,
    available_ram: getAvailableRAM(node),
    available_storage: getAvailableStorage(node),
    storage: node.total_resources?.sru ? Math.round(node.total_resources.sru / (1024 * 1024 * 1024)) : 0,
    price_usd: typeof node.price_usd === 'number' ? node.price_usd : null,
    discount_price: typeof node.discount_price === 'number' ? node.discount_price : null,
    gpu: (node.num_gpu && node.num_gpu > 0) || (node.gpus && node.gpus.length > 0),
    locationString: node.country + (node.city ? ', ' + node.city : ''),
    country: node.country,
    city: node.city,
    status: node.status,
    healthy: node.healthy,
    rentable: node.rentable,
    rented: node.rented,
    rentedByTwinId: node.rentedByTwinId,
    dedicated: node.dedicated,
    extraFee: node.extraFee,
    certificationType: node.certificationType,
  };
}

type ResourceKey = 'cru' | 'sru' | 'mru';

function getResourceValue(node: RentedNode, resourceKey: ResourceKey, used: boolean = false): number {
  if (!node) return 0;
  const resources = used ? node.used_resources : node.total_resources;
  const fallbackResources = used ? null : node.resources;
  if (resources && typeof resources[resourceKey] === 'number') {
    if (resourceKey === 'mru' || resourceKey === 'sru') {
      return Math.round(resources[resourceKey] / (1024 * 1024 * 1024));
    }
    return resources[resourceKey];
  }
  if (fallbackResources) {
    if (resourceKey === 'mru' && typeof fallbackResources.memory === 'number') {
      return fallbackResources.memory;
    }
    if (resourceKey === 'sru' && typeof fallbackResources.storage === 'number') {
      return fallbackResources.storage;
    }
    if (resourceKey === 'cru' && typeof fallbackResources.cpu === 'number') {
      return fallbackResources.cpu;
    }
  }
  return 0;
}

export function getTotalCPU(node: RentedNode): number {
  return getResourceValue(node, 'cru', false);
}
export function getUsedCPU(node: RentedNode): number {
  return getResourceValue(node, 'cru', true);
}
export function getAvailableCPU(node: RentedNode): number {
  if (!node) return 0;
  return getTotalCPU(node);
}
export function getTotalRAM(node: RentedNode): number {
  return getResourceValue(node, 'mru', false);
}
export function getUsedRAM(node: RentedNode): number {
  return getResourceValue(node, 'mru', true);
}
export function getAvailableRAM(node: RentedNode): number {
  if (!node) return 0;
  return Math.max(getTotalRAM(node) - getUsedRAM(node), 0);
}
export function getTotalStorage(node: RentedNode): number {
  return getResourceValue(node, 'sru', false);
}
export function getUsedStorage(node: RentedNode): number {
  return getResourceValue(node, 'sru', true);
}
export function getAvailableStorage(node: RentedNode): number {
  if (!node) return 0;
  return Math.max(getTotalStorage(node) - getUsedStorage(node), 0);
}
