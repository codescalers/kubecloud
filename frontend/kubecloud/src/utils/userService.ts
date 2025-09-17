import { WorkflowStatus } from '@/types/ewf'
import { api, createWorkflowStatusChecker, type ApiError } from './api'
import type { ApiResponse } from './authService'
import type { ChargeBalanceResponse } from './stripeService'
import { useNotificationStore } from '@/stores/notifications'
import type { NodeFilters } from '@/composables/useNodes'

export interface ReserveNodeRequest {
  // Add any required fields if needed
}

export interface ReserveNodeResponse {
  workflow_id: string
  node_id: number
  email: string
}
export interface UnreserveNodeResponse {
  workflow_id: string
  contract_id: number,
  email: string,
}
export interface ChargeBalanceRequest {
  card_type: string
  payment_method_id: string
  amount: number
}

export interface RedeemVoucherResponse {
  amount: number,
  email: string,
  voucher_code: string,
  workflow_id: string
}

export interface Node {
  id: number
  node_id: number
  farm_id: number
  twin_id: number
  country: string
  city: string
  latitude: number
  longitude: number
  created: number
  farming_policy_id: number
  interfaces: any[]
  secure: boolean
  virtualized: boolean
  serial_number: string
  created_at: number
  updated_at: number
  location_id: string
  power: {
    state: string
    target: string
  }
  public_config: {
    ipv4: string
    ipv6: string
    gw4: string
    gw6: string
    domain: string
  }
  public_ips: any[]
  resources: {
    hru: number
    sru: number
    cru: number
    mru: number
  }
  location: {
    country: string
    city: string
    latitude: number
    longitude: number
  }
  status: string
  healthy: boolean
  rentable: boolean
  rented: boolean
  rented_by: number
  rent_contract_id: number
  certification_type: string
  dedicated: boolean
  grid_version: number
}

export interface NodesResponse {
  total: number
  nodes: Node[]
}

export interface UserInvoice {
  id: number
  user_id: number
  total: number
  nodes: any[]
  tax: number
  created_at: string
}

export interface SshKey {
  ID: number
  name: string
  public_key: string
  created_at: string
  updated_at: string
}

export interface AddSshKeyRequest {
  name: string
  public_key: string
}

export interface TaskResponse {
  task_id: string;
  status: string;
  message: string;
  created_at: string;
}

export interface PendingRecord {
  created_at: string;
  id: number;
  tft_amount: number;
  transferred_tft_amount: number;
  transferred_usd_amount: number;
  updated_at: string;
  usd_amount: number;
  user_id: number;
  username: string;
  transfer_mode: string;
}

export interface TwinResponse {
  public_key: string;
  account_id: string;
  relay: string;
  twin_id: number;
}

export class UserService {
  async listNodes(filters?: NodeFilters) {
    const queryParams = new URLSearchParams()
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, String(value))
        }
      })
    }

    const endpoint = `${filters?.rentable ? '/v1/user/nodes/rentable' : '/v1/user/nodes'}${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return api.get<NodesResponse>(endpoint, {
      requiresAuth: true,
      showNotifications: false // Don't show notifications for node listing
    })
  }

  // Reserve a node
  async reserveNode(nodeId: number, data: ReserveNodeRequest = {}) {
    const response = await api.post<ApiResponse<ReserveNodeResponse>>(
      `/v1/user/nodes/${nodeId}`,
      data,
      { requiresAuth: true, showNotifications: true },
    )
    try {
      await this.trackNodeStatus(nodeId, 'rented')
    } catch (error) {
      useNotificationStore().error('Node reservation error', 'Failed to reserve node')
      throw new Error('Failed to reserve node')
    }
  }

  // List reserved nodes
  async listReservedNodes() {
    return api.get('/v1/user/nodes/rented', { requiresAuth: true })
  }

  // Unreserve a node
  async unreserveNode(contractId: string, nodeId: number) {
    const response = await api.delete<ApiResponse<UnreserveNodeResponse>>(
      `/v1/user/nodes/unreserve/${contractId}`,
      { requiresAuth: true, showNotifications: true },
    )
    try {
      await this.trackNodeStatus(nodeId, 'rentable')
    } catch (error) {
      useNotificationStore().error('Node unreservation error', 'Failed to verify node status')
    }
  }

  // Charge balance
  async chargeBalance(data: ChargeBalanceRequest) {
   await api.post<ApiResponse<ChargeBalanceResponse>>('/v1/user/balance/charge', data, {
     requiresAuth: true,
     showNotifications: true,
   })
  }

  // Create a PaymentIntent (to be implemented)
  // Replace this stub with a real API call to your backend that returns { clientSecret: string }
  async createPaymentIntent({ amount }: { amount: number }): Promise<{ clientSecret: string }> {
    throw new Error('UserService.createPaymentIntent is not implemented. Please implement this method to call your backend and return { clientSecret: string }.')
  }

  // List all invoices for the current user
  async listUserInvoices(): Promise<UserInvoice[]> {
    const response = await api.get<{ data: { invoices: UserInvoice[] } }>(
      '/v1/user/invoice',
      { requiresAuth: true, showNotifications: true, errorMessage: 'Failed to load invoices' }
    )
    return response.data.data.invoices
  }

  // Download a specific invoice as PDF
  async downloadInvoice(invoiceId: number): Promise<Blob> {
    // The backend should return application/pdf for this endpoint
    const response = await api.get(
      `/v1/user/invoice/${invoiceId}`,
      { requiresAuth: true, errorMessage: 'Failed to download invoice' }
    )
    return response.data as Blob
  }

  // Redeem a voucher
  async redeemVoucher(voucherCode: string) {
    await api.put<ApiResponse<RedeemVoucherResponse>>(
      `/v1/user/redeem/${voucherCode}`,
      {},
      {
        requiresAuth: true,
        showNotifications: true,
        errorMessage: 'Failed to redeem voucher',
      },
    )
  }

  // Fetch the user's current balance
  async fetchBalance(): Promise<{balance: number, pending_balance: number}> {
    try {
    const response = await api.get<{ data: { balance_usd: number, debt_usd: number, pending_balance_usd: number } }>(
      '/v1/user/balance',
      { requiresAuth: true, showNotifications: false }
    )
    const { balance_usd, debt_usd, pending_balance_usd } = response.data.data
    return {balance: (balance_usd || 0) - (debt_usd || 0), pending_balance: pending_balance_usd || 0}
  }catch(e){
    if (!(e as ApiError)?.silent) {
      useNotificationStore().error(
        'Error',
        'Failed to fetch balance',
      )
    }
    return {balance: 0, pending_balance: 0}
  }
  }

  // List all SSH keys for the current user
  async listSshKeys(): Promise<SshKey[]> {
    const response = await api.get<ApiResponse<SshKey[]>>('/v1/user/ssh-keys', {
      requiresAuth: true,
      showNotifications: false
    })
    return response.data.data
  }

  // Add a new SSH key
  async addSshKey(data: AddSshKeyRequest): Promise<SshKey> {
    const response = await api.post<ApiResponse<SshKey>>('/v1/user/ssh-keys', data, {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to add SSH key',
    })
    return response.data.data
  }

  // Delete an SSH key by ID
  async deleteSshKey(id: number): Promise<void> {
    await api.delete(`/v1/user/ssh-keys/${id}`, {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to delete SSH key',
    })
  }
  // Add node to deployment
  async addNodeToDeployment(deploymentName: string, clusterPayload: { name: string, nodes: any[] }) {
    return api.post<TaskResponse>(`/v1/deployments/${deploymentName}/nodes`, clusterPayload, { requiresAuth: true, showNotifications: true });
  }

  // Remove node from deployment
  async removeNodeFromDeployment(deploymentName: string, nodeName: string) {
    return api.delete(`/v1/deployments/${deploymentName}/nodes/${nodeName}`, { requiresAuth: true, showNotifications: true })
  }

  // List all pending records
  async listPendingRecords(): Promise<PendingRecord[]> {
    const response = await api.get<ApiResponse<PendingRecord[]>>('/v1/user/pending-records', {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load payments'
    })
    return response.data.data
  }

  async listUserPendingRecords(): Promise<PendingRecord[]> {
    const response = await api.get<{ data: { pending_records: PendingRecord[] } }>(`/v1/user/pending-records`, {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to load payments'
    })
    return response.data.data.pending_records
  }

  // Fetch twin account info
  async getTwinAccount(twinId: number): Promise<TwinResponse> {
    const response = await api.get<ApiResponse<TwinResponse>>(`/v1/twins/${twinId}/account`, {
      requiresAuth: true,
      showNotifications: false
    })
    return response.data.data
  }

  private async trackNodeStatus(nodeId: number, targetStatus: "rented" | "rentable", maxAttempts: number = 20, interval: number = 5000) {
    await new Promise((resolve, reject) => {
      let attempts = 0
      const intervalId = setInterval(async () => {
        const nodes = await this.listReservedNodes()

        const node =(nodes?.data as {data: {nodes: {nodeId: number}[]}}).data.nodes.find((n) => n.nodeId === nodeId)
        if((targetStatus === "rented" && node) || (targetStatus === "rentable" && !node)){
          clearInterval(intervalId)
          resolve(true)
        }
        attempts++
        if(attempts >= maxAttempts){
          clearInterval(intervalId)
          console.error(`Node status not reached after ${maxAttempts * interval / 1000} seconds`)
          reject()
        }
      }, interval)
    })
  }
}

export const userService = new UserService()
