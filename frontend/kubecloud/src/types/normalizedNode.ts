export interface NormalizedNode {
  nodeId: number;
  cpu: number; // vCPU
  ram: number; // GB
  storage: number; // GB
  available_ram?: number; // GB
  available_storage?: number; // GB
  price_usd: number | null;
  discount_price: number | null;
  gpu: boolean;
  locationString: string;
  country: string;
  city: string;
  status: string;
  healthy: boolean;
  rentable: boolean;
  rented: boolean;
  rentedByTwinId?: number;
  dedicated: boolean;
  certificationType: string;
  extraFee: number;
  farmId: number;
  twinId: number;
  // Add any other UI fields needed
}
