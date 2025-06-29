// Development utilities for debugging and testing

export const DEV_CONFIG = {
  enabled: import.meta.env.DEV,
  mockDelay: parseInt(import.meta.env.VITE_MOCK_DELAY || '1000'),
  enableLogging: import.meta.env.VITE_ENABLE_DEBUG_MODE === 'true',
  enablePerformanceMonitoring: import.meta.env.VITE_ENABLE_DEBUG_MODE === 'true',
}

// Enhanced console logging for development
export const devLog = {
  info: (message: string, data?: any) => {
    if (DEV_CONFIG.enableLogging) {
      console.log(`%c[DEV] ${message}`, 'color: #3B82F6; font-weight: bold;', data || '')
    }
  },
  
  success: (message: string, data?: any) => {
    if (DEV_CONFIG.enableLogging) {
      console.log(`%c[DEV] ✅ ${message}`, 'color: #10B981; font-weight: bold;', data || '')
    }
  },
  
  warning: (message: string, data?: any) => {
    if (DEV_CONFIG.enableLogging) {
      console.warn(`%c[DEV] ⚠️ ${message}`, 'color: #F59E0B; font-weight: bold;', data || '')
    }
  },
  
  error: (message: string, error?: any) => {
    if (DEV_CONFIG.enableLogging) {
      console.error(`%c[DEV] ❌ ${message}`, 'color: #EF4444; font-weight: bold;', error || '')
    }
  },
  
  group: (label: string, fn: () => void) => {
    if (DEV_CONFIG.enableLogging) {
      console.group(`%c[DEV] ${label}`, 'color: #8B5CF6; font-weight: bold;')
      fn()
      console.groupEnd()
    }
  },
  
  table: (data: any) => {
    if (DEV_CONFIG.enableLogging) {
      console.table(data)
    }
  }
}

// Performance monitoring
export const performanceMonitor = {
  start: (label: string) => {
    if (DEV_CONFIG.enablePerformanceMonitoring) {
      performance.mark(`${label}-start`)
    }
  },
  
  end: (label: string) => {
    if (DEV_CONFIG.enablePerformanceMonitoring) {
      performance.mark(`${label}-end`)
      performance.measure(label, `${label}-start`, `${label}-end`)
      
      const measure = performance.getEntriesByName(label)[0]
      devLog.info(`Performance: ${label}`, `${measure.duration.toFixed(2)}ms`)
    }
  },
  
  measure: async <T>(label: string, fn: () => Promise<T>): Promise<T> => {
    performanceMonitor.start(label)
    try {
      const result = await fn()
      return result
    } finally {
      performanceMonitor.end(label)
    }
  }
}

// Mock data utilities
export const mockDataUtils = {
  // Generate random data for testing
  generateRandomUser: () => ({
    id: Math.floor(Math.random() * 1000),
    name: `User ${Math.floor(Math.random() * 1000)}`,
    email: `user${Math.floor(Math.random() * 1000)}@example.com`,
    role: ['admin', 'user', 'developer'][Math.floor(Math.random() * 3)] as 'admin' | 'user' | 'developer',
    avatar: `https://i.pravatar.cc/150?img=${Math.floor(Math.random() * 70)}`,
    createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
    lastLogin: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString(),
  }),
  
  generateRandomCluster: () => ({
    id: `cluster-${Math.floor(Math.random() * 1000)}`,
    name: `Cluster ${Math.floor(Math.random() * 1000)}`,
    status: ['running', 'stopped', 'starting', 'stopping'][Math.floor(Math.random() * 4)] as any,
    region: ['us-west-2', 'us-east-1', 'eu-west-1', 'ap-southeast-1'][Math.floor(Math.random() * 4)],
    nodes: Math.floor(Math.random() * 10) + 1,
    cpu: `${Math.floor(Math.random() * 32) + 4} cores`,
    memory: `${Math.floor(Math.random() * 128) + 16} GB`,
    storage: `${Math.floor(Math.random() * 1000) + 100} GB`,
    createdAt: new Date(Date.now() - Math.random() * 60 * 24 * 60 * 60 * 1000).toISOString(),
    lastUpdated: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString(),
    cost: Math.random() * 500 + 50,
    tags: ['production', 'staging', 'development', 'testing'].slice(0, Math.floor(Math.random() * 3) + 1),
  }),
  
  // Simulate network conditions
  simulateNetworkDelay: (min: number = 500, max: number = 2000) => {
    const delay = Math.random() * (max - min) + min
    return new Promise(resolve => setTimeout(resolve, delay))
  },
  
  simulateNetworkError: (errorRate: number = 0.1) => {
    if (Math.random() < errorRate) {
      throw new Error('Simulated network error')
    }
  }
}

// Local storage utilities for development
export const devStorage = {
  set: (key: string, value: any) => {
    if (DEV_CONFIG.enabled) {
      localStorage.setItem(`dev_${key}`, JSON.stringify(value))
      devLog.info(`Dev storage set: ${key}`, value)
    }
  },
  
  get: <T>(key: string, defaultValue?: T): T | null => {
    if (DEV_CONFIG.enabled) {
      const stored = localStorage.getItem(`dev_${key}`)
      if (stored) {
        try {
          const value = JSON.parse(stored)
          devLog.info(`Dev storage get: ${key}`, value)
          return value
        } catch {
          devLog.error(`Dev storage parse error: ${key}`)
        }
      }
    }
    return defaultValue || null
  },
  
  remove: (key: string) => {
    if (DEV_CONFIG.enabled) {
      localStorage.removeItem(`dev_${key}`)
      devLog.info(`Dev storage removed: ${key}`)
    }
  },
  
  clear: () => {
    if (DEV_CONFIG.enabled) {
      const keys = Object.keys(localStorage).filter(key => key.startsWith('dev_'))
      keys.forEach(key => localStorage.removeItem(key))
      devLog.info('Dev storage cleared')
    }
  }
}

// Component debugging utilities
export const componentDebug = {
  // Track component lifecycle
  trackComponent: (componentName: string) => {
    if (DEV_CONFIG.enableLogging) {
      devLog.info(`Component mounted: ${componentName}`)
      
      return {
        unmount: () => devLog.info(`Component unmounted: ${componentName}`),
        update: (props?: any) => devLog.info(`Component updated: ${componentName}`, props),
      }
    }
    return { unmount: () => {}, update: () => {} }
  },
  
  // Track reactive updates
  trackReactive: <T>(name: string, value: T) => {
    if (DEV_CONFIG.enableLogging) {
      devLog.info(`Reactive update: ${name}`, value)
    }
    return value
  }
}

// API debugging utilities
export const apiDebug = {
  // Mock API response interceptor
  interceptResponse: (response: any, endpoint: string) => {
    if (DEV_CONFIG.enableLogging) {
      devLog.group(`API Response: ${endpoint}`, () => {
        devLog.info('Status', response.status)
        devLog.info('Data', response.data)
        devLog.table(response.data)
      })
    }
    return response
  },
  
  // Mock API error interceptor
  interceptError: (error: any, endpoint: string) => {
    if (DEV_CONFIG.enableLogging) {
      devLog.error(`API Error: ${endpoint}`, error)
    }
    return error
  }
}

// Export development utilities
export const devUtils = {
  log: devLog,
  performance: performanceMonitor,
  mock: mockDataUtils,
  storage: devStorage,
  component: componentDebug,
  api: apiDebug,
  config: DEV_CONFIG,
}

// Auto-initialize development tools
if (DEV_CONFIG.enabled) {
  devLog.info('Development tools initialized')
  
  // Add global dev utils for console access
  if (typeof window !== 'undefined') {
    (window as any).devUtils = devUtils
    devLog.info('Dev utils available globally as window.devUtils')
  }
} 