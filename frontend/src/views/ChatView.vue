<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useChatStore } from '@/stores/chat'
import { useContactsStore } from '@/stores/contacts'
import { useAuthStore } from '@/stores/auth'
import ChatWindow from '@/components/chat/ChatWindow.vue'
import Avatar from '@/components/common/Avatar.vue'
import type { Conversation, Message } from '@/types/chat'

const chat = useChatStore()
const contacts = useContactsStore()
const auth = useAuthStore()

const convs = computed(() => chat.conversations)
const searchKeyword = ref('')
const isSearching = computed(() => searchKeyword.value.trim().length > 0)

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

// ── 搜索：联系人昵称匹配 ──
const contactMatches = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return []
  return convs.value.filter((c) => c.name.toLowerCase().includes(kw))
})

// ── 搜索：聊天消息内容匹配（遍历已缓存的消息）──
interface MsgMatch {
  conv: Conversation
  msg: Message
  snippet: string   // 含关键词的截取片段
}
const messageMatches = computed<MsgMatch[]>(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return []
  const results: MsgMatch[] = []
  for (const [convId, msgs] of Object.entries(chat.messagesCache)) {
    const conv = convs.value.find((c) => c.id === convId)
    if (!conv) continue
    // 只取每个会话里第一条匹配的消息展示
    for (const msg of [...msgs].reverse()) {
      if (msg.is_recalled || msg.msg_type !== 'text') continue
      const raw = typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content)
      let text = ''
      try { text = (JSON.parse(raw) as { text: string }).text } catch { continue }
      if (!text.toLowerCase().includes(kw)) continue
      // 截取关键词前后各 15 字符作为摘要
      const idx = text.toLowerCase().indexOf(kw)
      const start = Math.max(0, idx - 15)
      const end = Math.min(text.length, idx + kw.length + 15)
      const snippet = (start > 0 ? '…' : '') + text.slice(start, end) + (end < text.length ? '…' : '')
      results.push({ conv, msg, snippet })
      break
    }
  }
  return results
})

// ── 折叠 / 展开 ──
const COLLAPSE_N = 3
const expanded = ref(new Set<string>())
watch(searchKeyword, () => { expanded.value = new Set() })

function toggle(key: string) {
  const next = new Set(expanded.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expanded.value = next
}
function isExpanded(key: string) { return expanded.value.has(key) }
function sliced<T>(key: string, list: T[]): T[] {
  return isExpanded(key) ? list : list.slice(0, COLLAPSE_N)
}

function clearSearch() {
  searchKeyword.value = ''
}

onMounted(() => {
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

function selectAndClear(convId: string) {
  chat.setActiveConv(convId)
  clearSearch()
}

function lastMsgPreview(conv: (typeof convs.value)[number]): string {
  const msg = conv.last_message
  if (!msg) return ''
  if (msg.is_recalled) {
    return msg.from_id === auth.user?.id ? '你撤回了一条消息' : `${msg.from?.nickname ?? '对方'} 撤回了一条消息`
  }
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

// 关键词高亮：将文本里的关键词包成 <mark>
function highlight(text: string): string {
  const kw = searchKeyword.value.trim()
  if (!kw) return text
  const re = new RegExp(kw.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
  return text.replace(re, (m) => `<mark>${m}</mark>`)
}
</script>

<template>
  <div class="view-layout">
    <!-- 会话列表 -->
    <aside class="conv-list">
      <div class="conv-list-header">
        <span class="conv-list-title">消息</span>
      </div>

      <!-- 搜索框 -->
      <div class="search-wrap">
        <div class="search-bar">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchKeyword"
            placeholder="搜索"
            @keydown.esc="clearSearch"
          />
          <button v-if="isSearching" class="search-clear" @click="clearSearch">✕</button>
        </div>
      </div>

      <!-- 搜索结果面板 -->
      <div v-if="isSearching" class="search-panel">
        <!-- 联系人 -->
        <template v-if="contactMatches.length > 0">
          <div class="search-section-label">
            联系人
            <span class="section-count">{{ contactMatches.length }}</span>
          </div>
          <div
            v-for="conv in sliced('contacts', contactMatches)"
            :key="conv.id"
            class="conv-item"
            :class="{ active: conv.id === chat.activeConvId }"
            @click="selectAndClear(conv.id)"
          >
            <Avatar
              :src="getAvatarSrc(conv.avatar)"
              :name="conv.name"
              :size="40"
              :status="contacts.isOnline(conv.target_id) ? 'online' : 'offline'"
              show-status
            />
            <div class="conv-info">
              <div class="conv-row1">
                <span class="conv-name" v-html="highlight(conv.name)" />
              </div>
            </div>
          </div>
          <button v-if="contactMatches.length > COLLAPSE_N" class="section-toggle" @click="toggle('contacts')">
            <template v-if="!isExpanded('contacts')">
              查看更多 {{ contactMatches.length - COLLAPSE_N }} 条
              <span class="chevron">›</span>
            </template>
            <template v-else>收起 <span class="chevron up">›</span></template>
          </button>
        </template>

        <!-- 聊天记录 -->
        <template v-if="messageMatches.length > 0">
          <div class="search-section-label">
            聊天记录
            <span class="section-count">{{ messageMatches.length }}</span>
          </div>
          <div
            v-for="item in sliced('messages', messageMatches)"
            :key="item.msg.id"
            class="conv-item"
            :class="{ active: item.conv.id === chat.activeConvId }"
            @click="selectAndClear(item.conv.id)"
          >
            <Avatar
              :src="getAvatarSrc(item.conv.avatar)"
              :name="item.conv.name"
              :size="40"
              :status="contacts.isOnline(item.conv.target_id) ? 'online' : 'offline'"
              show-status
            />
            <div class="conv-info">
              <div class="conv-row1">
                <span class="conv-name">{{ item.conv.name }}</span>
                <span class="conv-time">{{ timeStr(item.msg.created_at) }}</span>
              </div>
              <div class="conv-row2">
                <span class="conv-preview msg-snippet" v-html="highlight(item.snippet)" />
              </div>
            </div>
          </div>
          <button v-if="messageMatches.length > COLLAPSE_N" class="section-toggle" @click="toggle('messages')">
            <template v-if="!isExpanded('messages')">
              查看更多 {{ messageMatches.length - COLLAPSE_N }} 条
              <span class="chevron">›</span>
            </template>
            <template v-else>收起 <span class="chevron up">›</span></template>
          </button>
        </template>

        <!-- 无结果 -->
        <div v-if="contactMatches.length === 0 && messageMatches.length === 0" class="conv-empty">
          未找到相关内容
        </div>
      </div>

      <!-- 正常会话列表 -->
      <div v-else class="conv-items">
        <div
          v-for="conv in convs"
          :key="conv.id"
          class="conv-item"
          :class="{ active: conv.id === chat.activeConvId }"
          @click="selectConv(conv.id)"
        >
          <Avatar
            :src="getAvatarSrc(conv.avatar)"
            :name="conv.name"
            :size="40"
            :status="conv.chat_type === 'private' ? (contacts.isOnline(conv.target_id) ? 'online' : 'offline') : null"
            :show-status="conv.chat_type === 'private'"
          />
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
  padding: 16px 16px 6px;
  flex-shrink: 0;
}

.conv-list-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

/* 搜索框 */
.search-wrap { padding: 0 10px 8px; flex-shrink: 0; }

.search-bar {
  height: 32px;
  background: #E8E8E8;
  border-radius: var(--radius-search);
  display: flex; align-items: center; padding: 0 10px; gap: 6px;
  transition: box-shadow 0.15s;
}
.search-bar:focus-within { background: white; box-shadow: 0 0 0 1px var(--qq-blue-primary); }
.search-icon { font-size: 13px; opacity: 0.5; flex-shrink: 0; }
.search-bar input { flex: 1; font-size: 13px; color: var(--text-primary); background: transparent; user-select: text; min-width: 0; }
.search-clear {
  flex-shrink: 0; width: 16px; height: 16px; border-radius: 50%;
  background: var(--text-tertiary); color: white; border: none;
  font-size: 9px; cursor: pointer; display: flex; align-items: center; justify-content: center;
  opacity: 0.7; transition: opacity 0.12s; padding: 0;
}
.search-clear:hover { opacity: 1; }

/* 搜索面板 */
.search-panel { flex: 1; overflow-y: auto; }

.search-section-label {
  font-size: 11px; color: var(--text-tertiary);
  padding: 8px 14px 4px; user-select: none;
}

/* 关键词高亮 */
.search-panel :deep(mark),
.msg-snippet :deep(mark) {
  background: transparent;
  color: var(--qq-blue-primary);
  font-weight: 600;
}

.section-count {
  margin-left: 4px;
  font-size: 11px;
  color: var(--text-tertiary);
}

.section-toggle {
  width: 100%;
  padding: 6px 14px;
  text-align: left;
  font-size: 12px;
  color: var(--text-secondary);
  background: none;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color 0.12s;
}
.section-toggle:hover { color: var(--qq-blue-primary); }

.chevron {
  display: inline-block;
  font-style: normal;
  transform: rotate(90deg);
  transition: transform 0.2s;
  line-height: 1;
}
.chevron.up { transform: rotate(-90deg); }

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
