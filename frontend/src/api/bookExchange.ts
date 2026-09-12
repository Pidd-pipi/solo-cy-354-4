import request from '../utils/request'
import type { BookExchange, PageResult } from '../types'

export function createBookExchange(data: { offer_book: string; want_book: string; description?: string }) {
  return request.post<never, { code: number; message: string; data: BookExchange }>('/book-exchanges', data)
}

export function listBookExchanges(params: { page?: number; page_size?: number }) {
  return request.get<never, { code: number; message: string; data: PageResult<BookExchange> }>('/book-exchanges', { params })
}

export function closeBookExchange(id: number) {
  return request.post<never, { code: number; message: string; data: BookExchange }>(`/book-exchanges/${id}/close`)
}
