export interface ValidationRule {
  required?: boolean
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
    if (rules.minLength && value < rules.minLength) {
      errors.push(`${fieldName} must be at least ${rules.minLength}`)
    }
    if (rules.maxLength && value > rules.maxLength) {
      errors.push(`${fieldName} must be no more than ${rules.maxLength}`)
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
        return 'Password must contain at least 8 characters, including uppercase, lowercase, number, and special character'
      }
      return true
    }
  },
  URL: { required: true, url: true },
  PHONE: { required: true, pattern: PATTERNS.PHONE },
  ALPHANUMERIC: { required: true, pattern: PATTERNS.ALPHANUMERIC },
  HEX_COLOR: { pattern: PATTERNS.HEX_COLOR },
  IP_ADDRESS: { pattern: PATTERNS.IP_ADDRESS }
}

// Utility functions
export const sanitizeInput = (input: string): string => {
  return input.trim().replace(/[<>]/g, '')
}

export const formatValidationErrors = (errors: string[]): string => {
  return errors.join('. ')
}

export const createValidationRule = (rule: ValidationRule): ValidationRule => {
  return rule
}

// Async validation support
export const validateAsync = async (
  field: FieldValidation,
  asyncValidator?: (value: any) => Promise<boolean | string>
): Promise<ValidationResult> => {
  const syncResult = validateField(field)
  
  if (!syncResult.isValid || !asyncValidator) {
    return syncResult
  }

  try {
    const asyncResult = await asyncValidator(field.value)
    if (asyncResult !== true) {
      return {
        isValid: false,
        errors: [...syncResult.errors, typeof asyncResult === 'string' ? asyncResult : 'Validation failed']
      }
    }
  } catch (error) {
    return {
      isValid: false,
      errors: [...syncResult.errors, 'Validation error occurred']
    }
  }

  return syncResult
} 