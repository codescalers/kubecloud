import { ref } from 'vue';

/**
 * Root filesystem size in GB
 */
export const ROOTFS = 5;
export interface VM {
  name: string;
  vcpu: number;
  ram: number;
  node: number | null;
  rootfs: number;
  disk: number;
  gpu: boolean;
  sshKeyIds: number[];
  publicIp: boolean;
  planetary: boolean;
}
export interface DeployClusterNode { id: number; label: string; totalCPU: number; totalRAM: number; hasGPU: boolean; location: string; }
export interface SshKey { ID: number; name: string; public_key: string; created_at: string; updated_at: string; }

export function useDeployCluster() {
  const masters = ref<VM[]>([]);
  const workers = ref<VM[]>([]);
  const availableSshKeys = ref<SshKey[]>([]);

  function addMaster() {
    if (masters.value.length < 3) {
      masters.value.push({
        name: `Master${masters.value.length + 1}`,
        vcpu: 2,
        ram: 4,
        node: null,
        rootfs: ROOTFS,
        disk: 25,
        gpu: false,
        sshKeyIds: availableSshKeys.value.length ? [availableSshKeys.value[0].ID] : [],
        publicIp: false,
        planetary: false,
      });
    }
  }
  function addWorker() {
    workers.value.push({
      name: `Worker${workers.value.length + 1}`,
      vcpu: 2,
      ram: 4,
      node: null,
      rootfs: ROOTFS,
      disk: 25,
      gpu: false,
      sshKeyIds: availableSshKeys.value.length ? [availableSshKeys.value[0].ID] : [],
      publicIp: false,
      planetary: false,
    });
  }
  function removeMaster(idx: number) {
    masters.value.splice(idx, 1);
  }
  function removeWorker(idx: number) {
    workers.value.splice(idx, 1);
  }

  return {
    masters, workers, availableSshKeys,
    addMaster, addWorker, removeMaster, removeWorker,
  };
} 