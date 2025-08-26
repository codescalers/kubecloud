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
  EMAIL: /^(?=.{1,254}$)(?=.{1,64}@)[a-zA-Z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_`{|}~-]+)*@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/,
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
  IP_ADDRESS: { pattern: PATTERNS.IP_ADDRESS },
  CREDIT_AMOUNT: {
    required: true,
    custom: (value: number) => {
      if (typeof value !== 'number' || isNaN(value)) {
        return 'Amount must be a valid number'
      }
      if (value <= 0) {
        return 'Amount must be greater than 0'
      }
      if (value > 10000) {
        return 'Amount cannot exceed $10,000'
      }
      if (!/^\d+(\.\d{1,2})?$/.test(value.toString())) {
        return 'Amount can have at most 2 decimal places'
      }
      return true
    }
  },
  CREDIT_MEMO: {
    required: true,
    minLength: 3,
    maxLength: 255,
    custom: (value: string) => {
      if (typeof value !== 'string') {
        return 'Memo must be a string'
      }
      const trimmed = value.trim()
      if (trimmed.length < 3) {
        return 'Memo must be at least 3 characters long'
      }
      if (trimmed.length > 255) {
        return 'Memo cannot exceed 255 characters'
      }
      if (!/^[a-zA-Z0-9\s\-_.,!?()]+$/.test(trimmed)) {
        return 'Memo contains invalid characters. Only letters, numbers, spaces, and basic punctuation are allowed'
      }
      return true
    }
  }
}

// Node validation constants
export const NODE_VALIDATION = {
  CPU: { min: 1, max: 256 },
  RAM: { min: 2, max: 256 }, // in GB
  STORAGE: { min: 10, max: 10000 }, // in GB
  ROOTFS: { min: 5, max: 100 } // in GB
}

// Credit operation validation functions
export const validateCreditAmount = (amount: number): string | undefined => {
  const validation = validateField({
    value: amount,
    rules: VALIDATION_RULES.CREDIT_AMOUNT,
    fieldName: 'Amount'
  })

  return validation.isValid ? undefined : validation.errors[0]
}

export const validateCreditMemo = (memo: string): string | undefined => {
  const validation = validateField({
    value: memo,
    rules: VALIDATION_RULES.CREDIT_MEMO,
    fieldName: 'Memo'
  })

  return validation.isValid ? undefined : validation.errors[0]
}

export const validateCreditForm = (amount: number, memo: string): { isValid: boolean; errors: { amount: string; memo: string } } => {
  const amountError = validateCreditAmount(amount)
  const memoError = validateCreditMemo(memo)

  return {
    isValid: !amountError && !memoError,
    errors: {
      amount: amountError || '',
      memo: memoError || ''
    }
  }
}

// Additional validation utilities for auth operations
export const validateVerificationCode = (code: string): string | undefined => {
  if (!code.trim()) {
    return 'Verification code is required'
  }
  if (code.length < 4 || code.length > 6) {
    return 'Verification code must be 4-6 digits'
  }
  if (!/^\d+$/.test(code)) {
    return 'Verification code must contain only numbers'
  }
  return undefined
}

export const validateEmail = (email: string): string | undefined => {
  if (!email.trim()) {
    return 'Email is required'
  }
  if (!PATTERNS.EMAIL.test(email.trim())) {
    return 'Please enter a valid email address'
  }
  return undefined
}

export const validatePasswordStrength = (password: string): { error: string | undefined; strength: 'weak' | 'medium' | 'strong' } => {
  if (!password) {
    return { error: 'Password is required', strength: 'weak' }
  }

  const validation = validateField({
    value: password,
    rules: VALIDATION_RULES.PASSWORD,
    fieldName: 'Password'
  })

  if (!validation.isValid) {
    return { error: validation.errors[0], strength: 'weak' }
  }

  // Determine strength
  let strength: 'weak' | 'medium' | 'strong' = 'medium'
  if (password.length >= 12 && /[A-Z].*[A-Z]/.test(password) && /[0-9].*[0-9]/.test(password)) {
    strength = 'strong'
  } else if (password.length < 10) {
    strength = 'weak'
  }

  return { error: undefined, strength }
}

export const validateConfirmPassword = (password: string, confirmPassword: string): string | undefined => {
  if (!confirmPassword) {
    return 'Please confirm your password'
  }
  if (password !== confirmPassword) {
    return 'Passwords do not match'
  }
  return undefined
}

// Unified validation function for node fields
export const validateNodeField = (value: any, fieldName: string, rules: ValidationRule): string | undefined => {
  const result = validateField({ value, rules, fieldName })
  return result.isValid ? undefined : result.errors[0]
}

const toVuetifyRule = (validator: (value: any) => string | undefined) => 
  (value: any): string | boolean => validator(value) || true

const createNameRule = (fieldName: string, minLength: number = 3, maxLength: number = 20) => 
  (value: any): string | boolean => {
    const result = validateNodeField(value, fieldName, {
      required: true,
      minLength,
      maxLength,
      pattern: PATTERNS.ALPHANUMERIC
    })
    return result || true
  }

const createNumberRule = (fieldName: string, min: number, max: number) => 
  (value: any): string | boolean => {
    const result = validateNodeField(value, fieldName, {
      required: true,
      min,
      max
    })
    return result || true
  }

const createCustomRule = (fieldName: string, customValidator: (val: any) => boolean | string) => 
  (value: any): string | boolean => {
    const result = validateNodeField(value, fieldName, {
      required: true,
      custom: customValidator
    })
    return result || true
  }

export const RULES = {
  nodeName: createNameRule('Name'),
  clusterName: createNameRule('Cluster name'),  
  ram: createNumberRule('RAM', NODE_VALIDATION.RAM.min, NODE_VALIDATION.RAM.max),
  storage: createNumberRule('Storage', NODE_VALIDATION.STORAGE.min, NODE_VALIDATION.STORAGE.max),  
  cpu: createNumberRule('CPU', NODE_VALIDATION.CPU.min, NODE_VALIDATION.CPU.max),
  email: toVuetifyRule(validateEmail),
  verificationCode: toVuetifyRule(validateVerificationCode),  
  creditAmount: toVuetifyRule(validateCreditAmount),
  creditMemo: toVuetifyRule(validateCreditMemo),
  
  // Add missing validation rules for auth views
  name: createNameRule('Name', 3, 64),
  password: (value: any): string | boolean => {
    const result = validateField({
      value,
      rules: VALIDATION_RULES.PASSWORD,
      fieldName: 'Password'
    })
    return result.isValid ? true : result.errors[0]
  },
  confirmPassword: (value: any, password: string): string | boolean => {
    if (!value) return 'Please confirm your password'
    if (value !== password) return 'Passwords do not match'
    return true
  },

  voucherCount: (value: any): string | boolean => {
    const num = Number(value)
    if (isNaN(num)) return 'Number of vouchers must be a valid number'
    if (!Number.isInteger(num)) return 'Number of vouchers must be a whole number'
    if (num < 1) return 'Number of vouchers must be at least 1'
    if (num > 1000) return 'Number of vouchers cannot exceed 1000'
    return true
  },
  voucherExpiry: (value: any): string | boolean => {
    const num = Number(value)
    if (isNaN(num)) return 'Expiry days must be a valid number'
    if (!Number.isInteger(num)) return 'Expiry days must be a whole number'
    if (num < 1) return 'Expiry days must be at least 1'
    if (num > 365) return 'Expiry days cannot exceed 365'
    return true
  },

  validateNode: (node: any): Record<string, string> => {
    const validations = [
      ['name', RULES.nodeName(node.name)],
      ['vcpu', RULES.cpu(node.vcpu)],
      ['ram', RULES.ram(node.ram)],
      ['disk', RULES.storage(node.disk)]
    ]
    
    return Object.fromEntries(
      validations
        .filter(([, error]) => error !== true)
        .map(([field, error]) => [field, error as string])
    )
  }
}

// Utility functions
export const sanitizeInput = (input: string): string => {
  return input.trim().replace(/[<>]/g, '')
}

export const formatValidationErrors = (errors: string[]): string => {
  return errors.join('. ')
}
