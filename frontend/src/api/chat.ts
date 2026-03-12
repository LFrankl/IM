import client from './client'
import type { Message, Conversation } from '@/types/chat'

export const chatApi = {
  listConversations() {
    return client.get<Conversation[]>('/api/conversations')
  },
  getMessages(userId: number, beforeId?: number, limit = 30) {
    return client.get<Message[]>(`/api/messages/${userId}`, {
      params: { before_id: beforeId || undefined, limit },
    })
  },
  markRead(fromUserId: number) {
    return client.put(`/api/messages/${fromUserId}/read`)
  },
}
