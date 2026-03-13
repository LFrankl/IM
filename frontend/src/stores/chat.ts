import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Conversation, Message } from '@/types/chat'
import { chatApi } from '@/api/chat'
import type { ApiResponse } from '@/api/client'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const activeConvId = ref<string | null>(null)
  const messagesCache = ref<Record<string, Message[]>>({})
  const loadingMsgs = ref(false)

  const activeConv = computed(() =>
    conversations.value.find((c) => c.id === activeConvId.value) ?? null,
  )

  const totalUnread = computed(() =>
    conversations.value.reduce((sum, c) => sum + c.unread_count, 0),
  )

  async function fetchConversations() {
    const res = await chatApi.listConversations()
    const body = res.data as unknown as ApiResponse<Conversation[]>
    conversations.value = body.data ?? []
  }

  async function fetchMessages(targetId: number, beforeId?: number) {
    const convId = `private:${targetId}`
    loadingMsgs.value = true
    try {
      const res = await chatApi.getMessages(targetId, beforeId)
      const body = res.data as unknown as ApiResponse<Message[]>
      const msgs = body.data ?? []
      if (beforeId) {
        // 上翻历史：追加到头部
        messagesCache.value[convId] = [...msgs, ...(messagesCache.value[convId] ?? [])]
      } else {
        messagesCache.value[convId] = msgs
      }
      return msgs
    } finally {
      loadingMsgs.value = false
    }
  }

  function setActiveConv(convId: string | null) {
    activeConvId.value = convId || null
  }

  function upsertConversation(conv: Conversation) {
    const idx = conversations.value.findIndex((c) => c.id === conv.id)
    if (idx >= 0) {
      conversations.value[idx] = conv
    } else {
      conversations.value.unshift(conv)
    }
  }

  /** 收到新消息（来自 WS 推送或自己发送确认）*/
  function receiveMessage(msg: Message) {
    const convId = msg.chat_type === 'private'
      ? `private:${Number(msg.from_id)}`
      : `group:${Number(msg.to_id)}`

    // 写入消息缓存
    if (!messagesCache.value[convId]) messagesCache.value[convId] = []
    // 去重
    if (!messagesCache.value[convId].find((m) => m.id === msg.id)) {
      messagesCache.value[convId] = [...messagesCache.value[convId], msg]
    }

    // 更新会话列表
    const conv = conversations.value.find((c) => c.id === convId)
    if (conv) {
      const unread = convId !== activeConvId.value ? conv.unread_count + 1 : conv.unread_count
      const updated = { ...conv, last_message: msg, updated_at: msg.created_at, unread_count: unread }
      conversations.value = [updated, ...conversations.value.filter((c) => c.id !== convId)]
    } else {
      fetchConversations()
    }
  }

  /** 自己发送成功确认（message_sent 事件）*/
  function confirmSent(msg: Message) {
    const convId = `private:${Number(msg.to_id)}`
    if (!messagesCache.value[convId]) messagesCache.value[convId] = []
    if (!messagesCache.value[convId].find((m) => m.id === msg.id)) {
      messagesCache.value[convId] = [...messagesCache.value[convId], msg]
    }
    const conv = conversations.value.find((c) => c.id === convId)
    if (conv) {
      conv.last_message = msg
      conv.updated_at = msg.created_at
      conversations.value = [conv, ...conversations.value.filter((c) => c.id !== convId)]
    } else {
      fetchConversations()
    }
  }

  function clearUnread(convId: string) {
    const idx = conversations.value.findIndex((c) => c.id === convId)
    if (idx >= 0) {
      conversations.value = conversations.value.map((c) =>
        c.id === convId ? { ...c, unread_count: 0 } : c
      )
    }
  }

  async function recallMessage(msgId: number) {
    await chatApi.recallMessage(msgId)
    applyRecall(msgId)
  }

  function applyRecall(msgId: number) {
    for (const convId in messagesCache.value) {
      messagesCache.value[convId] = messagesCache.value[convId].map((m) =>
        m.id === msgId ? { ...m, is_recalled: true } : m
      )
    }
  }

  return {
    conversations,
    activeConvId,
    activeConv,
    messagesCache,
    loadingMsgs,
    totalUnread,
    fetchConversations,
    fetchMessages,
    setActiveConv,
    upsertConversation,
    receiveMessage,
    confirmSent,
    clearUnread,
    recallMessage,
    applyRecall,
  }
})
