import request from '../utils/request'
import type { Conversation, Message } from '../types'

export function createConversation(product_id: number) {
  return request.post<never, { code: number; message: string; data: Conversation }>('/conversations', { product_id })
}

export function listMyConversations() {
  return request.get<never, { code: number; message: string; data: Conversation[] }>('/conversations/me')
}

export function listMessages(conversationId: number) {
  return request.get<never, { code: number; message: string; data: Message[] }>(`/conversations/${conversationId}/messages`)
}

export function sendMessage(conversationId: number, content: string) {
  return request.post<never, { code: number; message: string; data: Message }>(`/conversations/${conversationId}/messages`, { content })
}
