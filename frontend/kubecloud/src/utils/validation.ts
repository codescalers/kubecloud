export interface ValidationRule {
  required?: boolean
  min?: number
  max?: number
  minLength?: number
  maxLength?: number
  pattern?: RegExp
  email?: boolean
  url?: boolean
  custom?: (value: any) => boolean | string
}

export interface ValidationResult {
  isValid: boolean
  errors: string[]
}

export interface FieldValidation {
  value: any
  rules: ValidationRule
  fieldName?: string
}

export class ValidationError extends Error {
  constructor(message: string, public field?: string) {
    super(message)
    this.name = 'ValidationError'
  }
}

// Common validation patterns
export const PATTERNS = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  URL: /^https?:\/\/.+/,
  PHONE: /^\+?[\d\s\-\(\)]+$/,
  PASSWORD: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
  ALPHANUMERIC: /^[a-zA-Z0-9]+$/,
  HEX_COLOR: /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/,
  IP_ADDRESS: /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
}

// Validation functions
export const validateField = (field: FieldValidation): ValidationResult => {
  const { value, rules, fieldName = 'Field' } = field
  const errors: string[] = []

  // Required validation
  if (rules.required && (!value || (typeof value === 'string' && value.trim() === ''))) {
    errors.push(`${fieldName} is required`)
  }

  // Skip other validations if value is empty and not required
  if (!value || (typeof value === 'string' && value.trim() === '')) {
    return { isValid: errors.length === 0, errors }
  }

  // Type-specific validations
  if (typeof value === 'string') {
    // Length validations
    if (rules.minLength && value.length < rules.minLength) {
      errors.push(`${fieldName} must be at least ${rules.minLength} characters`)
    }
    if (rules.maxLength && value.length > rules.maxLength) {
      errors.push(`${fieldName} must be no more than ${rules.maxLength} characters`)
    }

    // Pattern validations
    if (rules.pattern && !rules.pattern.test(value)) {
      errors.push(`${fieldName} format is invalid`)
    }

    // Email validation
    if (rules.email && !PATTERNS.EMAIL.test(value)) {
      errors.push(`${fieldName} must be a valid email address`)
    }

    // URL validation
    if (rules.url && !PATTERNS.URL.test(value)) {
      errors.push(`${fieldName} must be a valid URL`)
    }
  }

  // Number validations
  if (typeof value === 'number') {
    if (rules.min !== undefined && value < rules.min) {
      errors.push(`${fieldName} must be at least ${rules.min}`)
    }
    if (rules.max !== undefined && value > rules.max) {
      errors.push(`${fieldName} must be no more than ${rules.max}`)
    }
  }

  // Custom validation
  if (rules.custom) {
    const customResult = rules.custom(value)
    if (customResult !== true) {
      errors.push(typeof customResult === 'string' ? customResult : `${fieldName} is invalid`)
    }
  }

  return { isValid: errors.length === 0, errors }
}

export const validateForm = (fields: Record<string, FieldValidation>): ValidationResult => {
  const allErrors: string[] = []
  let isValid = true

  for (const [fieldName, field] of Object.entries(fields)) {
    const result = validateField({
      ...field,
      fieldName: field.fieldName || fieldName
    })

    if (!result.isValid) {
      isValid = false
      allErrors.push(...result.errors)
    }
  }

  return { isValid, errors: allErrors }
}

// Common validation rules
export const VALIDATION_RULES = {
  REQUIRED: { required: true },
  EMAIL: { required: true, email: true },
  PASSWORD: {
    required: true,
    minLength: 8,
    pattern: PATTERNS.PASSWORD,
    custom: (value: string) => {
      if (!PATTERNS.PASSWORD.test(value)) {
        return 'Password must contain at least 8 characters, including uppercase, lowercase, number, and special character (@$!%*?&)'
      }
      return true
    }
  },
  URL: { required: true, url: true },
  PHONE: { required: true, pattern: PATTERNS.PHONE },
  ALPHANUMERIC: { required: true, pattern: PATTERNS.ALPHANUMERIC },
  CLUSTER_NAME: {
    required: true,
    minLength: 3,
    maxLength: 20,
    pattern: PATTERNS.ALPHANUMERIC
  },
  NODE_NAME: {
    required: true,
    minLength: 3,
    maxLength: 20,
    pattern: PATTERNS.ALPHANUMERIC
  },
  HEX_COLOR: { pattern: PATTERNS.HEX_COLOR },
  IP_ADDRESS: { pattern: PATTERNS.IP_ADDRESS }
}

// Node validation constants
export const NODE_VALIDATION = {
  CPU: { min: 1, max: 255 },
  RAM: { min: 2, max: 256 }, // in GB
  STORAGE: { min: 10, max: 10000 }, // in GB
  ROOTFS: { min: 5, max: 100 } // in GB
}

// Unified validation function for node fields
export const validateNodeField = (value: any, fieldName: string, rules: ValidationRule): string | undefined => {
  const result = validateField({ value, rules, fieldName })
  return result.isValid ? undefined : result.errors[0]
}

// Vuetify-compatible validation rules
export const RULES = {
  nodeName: (value: any) => validateNodeField(value, 'Name', {
    required: true,
    minLength: 3,
    maxLength: 20,
    pattern: PATTERNS.ALPHANUMERIC
  }),

  cpu: (value: any) => validateNodeField(value, 'CPU', {
    required: true,
    min: 1,
    max: 255,
    custom: (val: any) => {
      const num = Number(val)
      if (isNaN(num)) return 'CPU must be a valid number'
      if (!Number.isInteger(num)) return 'CPU must be a whole number (no decimals)'
      return true
    }
  }),

  ram: (value: any) => validateNodeField(value, 'RAM', {
    required: true,
    min: 2,
    max: 256
  }),

  storage: (value: any) => validateNodeField(value, 'Storage', {
    required: true,
    min: 10,
    max: 10000
  }),

  clusterName: (value: any) => validateNodeField(value, 'Cluster name', {
    required: true,
    minLength: 3,
    maxLength: 20,
    pattern: PATTERNS.ALPHANUMERIC
  }),

  validateNode: (node: any): Record<string, string> => {
    const errors: Record<string, string> = {}

    const nameError = RULES.nodeName(node.name)
    if (nameError) errors.name = nameError

    const cpuError = RULES.cpu(node.vcpu)
    if (cpuError) errors.vcpu = cpuError

    const ramError = RULES.ram(node.ram)
    if (ramError) errors.ram = ramError

    const storageError = RULES.storage(node.disk)
    if (storageError) errors.disk = storageError

    return errors
  }
}

// Vuetify-compatible validation rules
export const VUETIFY_RULES = {
  nodeName: (value: any): string | boolean => {
    const result = RULES.nodeName(value);
    return result || true;
  },
  cpu: (value: any): string | boolean => {
    const result = RULES.cpu(value);
    return result || true;
  },
  ram: (value: any): string | boolean => {
    const result = RULES.ram(value);
    return result || true;
  },
  storage: (value: any): string | boolean => {
    const result = RULES.storage(value);
    return result || true;
  },
  clusterName: (value: any): string | boolean => {
    const result = RULES.clusterName(value);
    return result || true;
  }
}

// Utility functions
export const sanitizeInput = (input: string): string => {
  return input.trim().replace(/[<>]/g, '')
}

export const formatValidationErrors = (errors: string[]): string => {
  return errors.join('. ')
}
