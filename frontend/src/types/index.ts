export interface User {
  id: number
  phone: string
  nickname: string
  avatar: string
  role: string
  campus: string
  credit_score: number
}

export interface LoginResponse {
  token: string
  user: User
}

export interface Product {
  id: number
  seller_id: number
  title: string
  description: string
  price: number
  category: string
  condition: string
  campus: string
  trade_location: string
  images: string
  status: string
  created_at: string
}

export interface Conversation {
  id: number
  product_id: number
  buyer_id: number
  seller_id: number
  created_at: string
}

export interface Message {
  id: number
  conversation_id: number
  sender_id: number
  content: string
  read: boolean
  created_at: string
}

export interface TradeOrder {
  id: number
  product_id: number
  buyer_id: number
  seller_id: number
  status: string
  buyer_confirmed_at: string | null
  seller_confirmed_at: string | null
  completed_at: string | null
  created_at: string
}

export interface Review {
  id: number
  trade_id: number
  reviewer_id: number
  reviewee_id: number
  rating: string
  content: string
  created_at: string
}

export interface BookExchange {
  id: number
  user_id: number
  offer_book: string
  want_book: string
  description: string
  status: string
  matched_id: number | null
  created_at: string
}

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}
