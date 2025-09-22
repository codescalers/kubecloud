export interface RawNode {
  id: string;
  nodeId: number;
  farmId: number;
  farmName: string;
  twinId: number;
  country: string;
  gridVersion: number;
  city: string;
  uptime: number;
  created: number;
  farmingPolicyId: number;
  updatedAt: number;
  total_resources: {
    cru: number;
    sru: number;
    hru: number;
    mru: number;
  };
  used_resources: {
    cru: number;
    sru: number;
    hru: number;
    mru: number;
  };
  location: {
    country: string;
    city: string;
    longitude: number;
    latitude: number;
  };
  publicConfig: {
    domain: string;
    gw4: string;
    gw6: string;
    ipv4: string;
    ipv6: string;
  };
  status: string;
  certificationType: string;
  dedicated: boolean;
  inDedicatedFarm: boolean;
  rentContractId: number;
  rented: boolean;
  rentable: boolean;
  rentedByTwinId: number;
  serialNumber: string;
  power: {
    state: string;
    target: string;
  };
  num_gpu: number;
  extraFee: number;
  healthy: boolean;
  dmi: {
    bios: {
      vendor: string;
      version: string;
    };
    baseboard: {
      manufacturer: string;
      product_name: string;
    };
    processor: Array<{
      version: string;
      thread_count: string;
    }>;
    memory: Array<{
      manufacturer: string;
      type: string;
    }>;
  };
  speed: {
    upload: number;
    download: number;
  };
  gpus: any[];
  price_usd: number;
  discount_price: number;
  farm_free_ips: number;
  features: string[];
} 