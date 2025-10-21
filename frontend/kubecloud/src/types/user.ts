export interface User {
  id: number
  username: string
  email: string
  admin: boolean
  verified: boolean
  updated_at: string
  balance_usd?: number
  pending_balance_usd?: number
  balance: number
  credit_card_balance: number
  credited_balance: number
  credit_card_balance_in_usd: number
  credited_balance_in_usd: number
  stripe_customer_id: string
  ssh_key: string
  debt: number
  debt_in_usd: number
  balance_in_tft: number
  sponsored: boolean
  account_address: string
  code: number
}
