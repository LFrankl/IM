<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import { useContactsStore } from '@/stores/contacts'
import ChatWindow from '@/components/chat/ChatWindow.vue'
import Avatar from '@/components/common/Avatar.vue'
import type { Conversation } from '@/types/chat'

const chat = useChatStore()
const contacts = useContactsStore()

const convs = computed(() => chat.conversations)

// 当从联系人页面跳转但会话列表里还没有记录时，用好友信息构造临时会话
const activeConv = computed<Conversation | null>(() => {
  if (!chat.activeConvId) return null
  if (chat.activeConv) return chat.activeConv

  const match = chat.activeConvId.match(/^private:(\d+)$/)
  if (!match) return null
  const targetId = parseInt(match[1])
  const friendship = contacts.friendships.find((f) => f.friend_id === targetId)
  if (!friendship) return null

  return {
    id: chat.activeConvId,
    chat_type: 'private',
    target_id: targetId,
    name: friendship.remark || friendship.friend?.nickname || '未知',
    avatar: friendship.friend?.avatar || '',
    last_message: null,
    unread_count: 0,
    updated_at: '',
  }
})

onMounted(() => {
  // 只有从联系人页"发消息"跳来时才保留 activeConvId，否则清空
  const state = history.state as { openConv?: string } | null
  if (!state?.openConv) {
    chat.setActiveConv('')
  }
  chat.fetchConversations()
})

onUnmounted(() => {
  chat.setActiveConv(null)
})

function selectConv(convId: string) {
  chat.setActiveConv(convId)
}

function lastMsgPreview(conv: (typeof convs.value)[number]): string {
  const msg = conv.last_message
  if (!msg) return ''
  if (msg.msg_type === 'text') {
    const c = typeof msg.content === 'string' ? JSON.parse(msg.content) : msg.content
    return (c as { text: string }).text
  }
  if (msg.msg_type === 'image') return '[图片]'
  if (msg.msg_type === 'file') return '[文件]'
  return ''
}

function timeStr(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  if (isToday) {
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  }
  return `${d.getMonth() + 1}/${d.getDate()}`
}

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}
</script>

<template>
  <div class="view-layout">
    <!-- 会话列表 -->
    <aside class="conv-list">
      <div class="conv-list-header">
        <span class="conv-list-title">消息</span>
      </div>

      <div class="conv-items">
        <div
          v-for="conv in convs"
          :key="conv.id"
          class="conv-item"
          :class="{ active: conv.id === chat.activeConvId }"
          @click="selectConv(conv.id)"
        >
          <!-- 头像 -->
          <Avatar
            :src="getAvatarSrc(conv.avatar)"
            :name="conv.name"
            :size="40"
            :status="conv.chat_type === 'private' ? (contacts.isOnline(conv.target_id) ? 'online' : 'offline') : null"
            :show-status="conv.chat_type === 'private'"
          />

          <!-- 信息 -->
          <div class="conv-info">
            <div class="conv-row1">
              <span class="conv-name">{{ conv.name }}</span>
              <span class="conv-time">{{ timeStr(conv.updated_at) }}</span>
            </div>
            <div class="conv-row2">
              <span class="conv-preview">{{ lastMsgPreview(conv) }}</span>
              <span v-if="conv.unread_count > 0" class="conv-badge">
                {{ conv.unread_count > 99 ? '99+' : conv.unread_count }}
              </span>
            </div>
          </div>
        </div>

        <div v-if="convs.length === 0" class="conv-empty">
          暂无会话
        </div>
      </div>
    </aside>

    <!-- 聊天窗口 -->
    <main class="chat-area">
      <ChatWindow v-if="activeConv" :conv="activeConv" :key="activeConv.id" />
      <div v-else class="no-chat">
        <div class="no-chat-icon">💬</div>
        <div class="no-chat-text">选择一个会话开始聊天</div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.view-layout {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

/* 会话列表 */
.conv-list {
  width: var(--list-width);
  flex-shrink: 0;
  background: var(--bg-list);
  border-right: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.conv-list-header {
  padding: 16px 16px 10px;
  flex-shrink: 0;
}

.conv-list-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.conv-items {
  flex: 1;
  overflow-y: auto;
}

.conv-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  cursor: pointer;
  transition: background 0.12s;
}

.conv-item:hover {
  background: var(--bg-hover);
}

.conv-item.active {
  background: var(--bg-active);
}

.conv-avatar {
  flex-shrink: 0;
}

.conv-info {
  flex: 1;
  min-width: 0;
}

.conv-row1 {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 3px;
}

.conv-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 140px;
}

.conv-time {
  font-size: 11px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.conv-row2 {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.conv-preview {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160px;
}

.conv-badge {
  background: var(--qq-red);
  color: white;
  font-size: 11px;
  font-weight: 600;
  border-radius: 10px;
  padding: 0 5px;
  min-width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.conv-empty {
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
  padding: 40px 0;
}

/* 聊天区 */
.chat-area {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.no-chat {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.no-chat-icon {
  font-size: 52px;
  opacity: 0.2;
}

.no-chat-text {
  font-size: 14px;
  color: var(--text-tertiary);
}
</style>
