import { api, type ApiResponse } from "./api";
import type { VM } from "../types/vms";



export function listVMs() : Promise<ApiResponse<VM[]>> {
  return api.get('/v1/deployments/vms', { requiresAuth: true })
}

export function getVM(id: number) : Promise<ApiResponse<VM>> {
  return api.get(`/v1/deployments/vms/${id}`, { requiresAuth: true })
}

export function deployVM(vmData: any) : Promise<ApiResponse<VM>> {
  return api.post('/v1/deployments/vms', vmData, { requiresAuth: true })
}

export function deleteVM(id: number) {
  return api.delete(`/v1/deployments/vms/${id}`, { requiresAuth: true })
}

export function deleteAllVMs() {
  return api.delete('/v1/deployments/vms', { requiresAuth: true })
}


