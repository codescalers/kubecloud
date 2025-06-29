import { ref } from 'vue'

// Mock data configuration
export const MOCK_CONFIG = {
  enabled: import.meta.env.VITE_ENABLE_MOCK_DATA === 'true',
  delay: parseInt(import.meta.env.VITE_MOCK_DELAY || '1000'),
  errorRate: 0.05, // 5% chance of error
}

// Simulate API delay
export const mockDelay = (ms: number = MOCK_CONFIG.delay) => 
  new Promise(resolve => setTimeout(resolve, ms))

// Simulate random errors
export const simulateError = () => 
  Math.random() < MOCK_CONFIG.errorRate

// Mock user data with proper types
export const mockUsers = [
  {
    id: 1,
    name: 'Admin User',
    email: 'admin@kubecloud.com',
    role: 'admin' as const,
    avatar: 'https://i.pravatar.cc/150?img=1',
    createdAt: '2024-01-15T10:00:00Z',
    lastLogin: '2024-03-15T14:30:00Z',
  },
  {
    id: 2,
    name: 'Regular User',
    email: 'user@kubecloud.com',
    role: 'user' as const,
    avatar: 'https://i.pravatar.cc/150?img=2',
    createdAt: '2024-02-01T09:00:00Z',
    lastLogin: '2024-03-14T16:45:00Z',
  },
  {
    id: 3,
    name: 'Developer User',
    email: 'dev@kubecloud.com',
    role: 'developer' as const,
    avatar: 'https://i.pravatar.cc/150?img=3',
    createdAt: '2024-02-15T11:00:00Z',
    lastLogin: '2024-03-15T10:20:00Z',
  }
]

// Mock cluster data with proper types
export const mockClusters = [
  {
    id: 'cluster-1',
    name: 'Production Cluster',
    status: 'running' as const,
    region: 'us-west-2',
    nodes: 5,
    cpu: '20 cores',
    memory: '80 GB',
    storage: '500 GB',
    createdAt: '2024-01-10T08:00:00Z',
    lastUpdated: '2024-03-15T12:00:00Z',
    cost: 245.50,
    tags: ['production', 'critical'],
  },
  {
    id: 'cluster-2',
    name: 'Staging Cluster',
    status: 'running' as const,
    region: 'us-east-1',
    nodes: 3,
    cpu: '12 cores',
    memory: '48 GB',
    storage: '200 GB',
    createdAt: '2024-02-01T10:00:00Z',
    lastUpdated: '2024-03-15T11:30:00Z',
    cost: 156.75,
    tags: ['staging', 'testing'],
  },
  {
    id: 'cluster-3',
    name: 'Development Cluster',
    status: 'stopped' as const,
    region: 'eu-west-1',
    nodes: 2,
    cpu: '8 cores',
    memory: '32 GB',
    storage: '100 GB',
    createdAt: '2024-02-15T14:00:00Z',
    lastUpdated: '2024-03-14T18:00:00Z',
    cost: 89.25,
    tags: ['development', 'temporary'],
  }
]

// Mock billing data
export const mockBilling = {
  currentMonth: {
    total: 1245.75,
    clusters: 890.50,
    storage: 245.25,
    bandwidth: 110.00,
  },
  previousMonth: {
    total: 1180.30,
    clusters: 845.75,
    storage: 230.55,
    bandwidth: 104.00,
  },
  usage: [
    { date: '2024-03-01', cost: 42.50 },
    { date: '2024-03-02', cost: 41.75 },
    { date: '2024-03-03', cost: 43.20 },
    { date: '2024-03-04', cost: 44.10 },
    { date: '2024-03-05', cost: 42.90 },
    { date: '2024-03-06', cost: 43.60 },
    { date: '2024-03-07', cost: 45.20 },
    { date: '2024-03-08', cost: 44.80 },
    { date: '2024-03-09', cost: 43.95 },
    { date: '2024-03-10', cost: 44.30 },
    { date: '2024-03-11', cost: 43.70 },
    { date: '2024-03-12', cost: 44.15 },
    { date: '2024-03-13', cost: 43.85 },
    { date: '2024-03-14', cost: 44.50 },
    { date: '2024-03-15', cost: 45.10 },
  ]
}

// Mock notifications
export const mockNotifications = [
  {
    id: 'notif-1',
    type: 'info',
    title: 'Cluster Update Available',
    message: 'A new version of Kubernetes is available for your clusters.',
    timestamp: '2024-03-15T10:30:00Z',
    read: false,
  },
  {
    id: 'notif-2',
    type: 'success',
    title: 'Backup Completed',
    message: 'Scheduled backup for Production Cluster completed successfully.',
    timestamp: '2024-03-15T08:00:00Z',
    read: true,
  },
  {
    id: 'notif-3',
    type: 'warning',
    title: 'High Resource Usage',
    message: 'Development Cluster is using 85% of allocated CPU.',
    timestamp: '2024-03-14T16:45:00Z',
    read: false,
  }
]

// Mock API responses
export const createMockResponse = <T>(data: T, success: boolean = true) => ({
  data,
  status: success ? 200 : 400,
  message: success ? 'Success' : 'Error occurred',
  timestamp: new Date().toISOString(),
})

// Mock API client for development
export class MockApiClient {
  private delay: number

  constructor(delay: number = MOCK_CONFIG.delay) {
    this.delay = delay
  }

  async get<T>(endpoint: string): Promise<T> {
    await mockDelay(this.delay)
    
    if (simulateError()) {
      throw new Error('Mock API error')
    }

    switch (endpoint) {
      case '/users':
        return createMockResponse(mockUsers) as T
      case '/clusters':
        return createMockResponse(mockClusters) as T
      case '/billing':
        return createMockResponse(mockBilling) as T
      case '/notifications':
        return createMockResponse(mockNotifications) as T
      default:
        throw new Error(`Mock endpoint not found: ${endpoint}`)
    }
  }

  async post<T>(endpoint: string, data?: any): Promise<T> {
    await mockDelay(this.delay)
    
    if (simulateError()) {
      throw new Error('Mock API error')
    }

    // Simulate successful creation
    return createMockResponse({ id: Date.now(), ...data }) as T
  }

  async put<T>(endpoint: string, data: any): Promise<T> {
    await mockDelay(this.delay)
    
    if (simulateError()) {
      throw new Error('Mock API error')
    }

    return createMockResponse(data) as T
  }

  async delete<T>(endpoint: string): Promise<T> {
    await mockDelay(this.delay)
    
    if (simulateError()) {
      throw new Error('Mock API error')
    }

    return createMockResponse({ success: true }) as T
  }
}

// Export mock client instance
export const mockApi = new MockApiClient() 