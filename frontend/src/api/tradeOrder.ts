import request from '../utils/request'
import type { PageResult, TradeOrder } from '../types'

export function createTradeOrder(product_id: number) {
  return request.post<never, { code: number; message: string; data: TradeOrder }>('/trade-orders', { product_id })
}

export function listMyOrders(params: { page?: number; page_size?: number }) {
  return request.get<never, { code: number; message: string; data: PageResult<TradeOrder> }>('/trade-orders/me', { params })
}

export function buyerConfirm(id: number) {
  return request.post<never, { code: number; message: string; data: TradeOrder }>(`/trade-orders/${id}/buyer-confirm`)
}

export function sellerConfirm(id: number) {
  return request.post<never, { code: number; message: string; data: TradeOrder }>(`/trade-orders/${id}/seller-confirm`)
}

export function cancelTradeOrder(id: number) {
  return request.post<never, { code: number; message: string; data: TradeOrder }>(`/trade-orders/${id}/cancel`)
}
