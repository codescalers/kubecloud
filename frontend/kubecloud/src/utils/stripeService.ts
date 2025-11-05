import { useNotificationStore } from '@/stores/notifications'
import { loadStripe, type Stripe } from '@stripe/stripe-js'

export interface PaymentMethod {
  id: string
  type: string
  card?: {
    brand: string
    last4: string
    exp_month: number
    exp_year: number
  }
  billing_details?: {
    name: string
    email: string
  }
}

export interface PaymentIntent {
  id: string
  client_secret: string
  amount: number
  currency: string
  status: string
}

export interface ChargeBalanceRequest {
  card_type: string
  payment_method_id: string
  amount: number
}

export interface ChargeBalanceResponse {
  email: string
  workflow_id: string
}

class StripeService {
  private stripe: Stripe | null = null
  private isInitialized = false

  async initialize(): Promise<void> {
    if (this.isInitialized) return

    const stripePublishableKey =
      (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_STRIPE_PUBLISHABLE_KEY) ||
      import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY
    if (!stripePublishableKey) {
      throw new Error('Stripe publishable key not found in environment variables')
    }

    this.stripe = await loadStripe(stripePublishableKey)
    this.isInitialized = true
  }

  async getStripe(): Promise<Stripe> {
    if (!this.isInitialized) {
      await this.initialize()
    }
    if (!this.stripe) {
      throw new Error('Stripe failed to initialize')
    }
    return this.stripe
  }

  // Create a token from card details (for legacy or backend expecting 'tok_' tokens)
  // Use this when your backend expects a token id (starts with 'tok_') instead of a payment method id (starts with 'pm_')
  async createToken(cardElement: any, billingDetails?: any): Promise<string> {
    const stripe = await this.getStripe()
    const { token, error } = await stripe.createToken(cardElement, billingDetails)

    if (!token) {
      const errorMessage = error?.message || 'Token creation failed'
      const notificationStore = useNotificationStore()
      notificationStore.error('Stripe Error', errorMessage)
      throw new Error(errorMessage)
    }

    return token.id // This will be a 'tok_...' string
  }

  // Create a payment method from card details (for PaymentMethod id 'pm_...')
  // Use this when your backend expects a payment method id (starts with 'pm_')
  async createPaymentMethod(cardElement: any, billingDetails?: any): Promise<PaymentMethod> {
    const stripe = await this.getStripe()

    const { paymentMethod, error } = await stripe.createPaymentMethod({
      type: 'card',
      card: cardElement,
      billing_details: billingDetails,
    })

    if (error) {
      throw new Error(error.message)
    }
    console.log({ paymentMethod })

    return paymentMethod as PaymentMethod
  }

  // Confirm a payment intent
  async confirmPayment(clientSecret: string, paymentMethodId: string): Promise<any> {
    const stripe = await this.getStripe()

    const result = await stripe.confirmPayment({
      clientSecret,
      confirmParams: {
        payment_method: paymentMethodId,
        return_url: `${window.location.origin}/dashboard/payment`,
      },
    })

    if (result.error) {
      throw new Error(result.error.message)
    }

    return result
  }

  // Note: These methods require server-side implementation
  // The Stripe client-side library doesn't support these operations directly
  // They should be implemented through your backend API

  // Get payment method details (requires backend)
  async getPaymentMethod(paymentMethodId: string): Promise<PaymentMethod> {
    throw new Error('getPaymentMethod requires backend implementation')
  }

  // List payment methods for a customer (requires backend)
  async listPaymentMethods(customerId: string): Promise<PaymentMethod[]> {
    throw new Error('listPaymentMethods requires backend implementation')
  }

  // Detach a payment method (requires backend)
  async detachPaymentMethod(paymentMethodId: string): Promise<void> {
    throw new Error('detachPaymentMethod requires backend implementation')
  }
}

export const stripeService = new StripeService()
