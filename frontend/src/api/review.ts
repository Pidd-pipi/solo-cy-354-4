import request from '../utils/request'
import type { Review } from '../types'

export function createReview(data: { trade_id: number; rating: string; content?: string }) {
  return request.post<never, { code: number; message: string; data: Review }>('/reviews', data)
}

export function listMyReviews() {
  return request.get<never, { code: number; message: string; data: Review[] }>('/reviews/me')
}
