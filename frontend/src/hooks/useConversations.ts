import { ref } from 'vue'
import { listMyConversations, listMessages, sendMessage } from '../api/conversation'
import type { Conversation, Message } from '../types'

export function useConversations() {
  const conversations = ref<Conversation[]>([])
  const messages = ref<Message[]>([])
  const activeId = ref<number | null>(null)

  async function loadConversations() {
    const res = await listMyConversations()
    conversations.value = res.data
  }

  async function open(convId: number) {
    activeId.value = convId
    const res = await listMessages(convId)
    messages.value = res.data
  }

  async function send(content: string) {
    if (activeId.value === null) return
    await sendMessage(activeId.value, content)
    await open(activeId.value)
  }

  return { conversations, messages, activeId, loadConversations, open, send }
}
