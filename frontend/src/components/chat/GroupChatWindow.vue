<script setup lang="ts">
import { ref, computed, watch, watchEffect } from 'vue'
import { useGroupStore } from '@/stores/group'
import { useAuthStore } from '@/stores/auth'
import { useWS } from '@/composables/useWS'
import { useUserCard } from '@/composables/useUserCard'
import { groupApi } from '@/api/group'
import ChatBubble from '@/components/chat/ChatBubble.vue'
import Avatar from '@/components/common/Avatar.vue'
import type { GroupWithMeta } from '@/stores/group'
import type { GroupMember } from '@/types/group'

const props = defineProps<{ group: GroupWithMeta }>()
const emit = defineEmits<{ kick: [userId: number] }>()

const store = useGroupStore()
const auth = useAuthStore()
const ws = useWS()

const msgListRef = ref<HTMLElement | null>(null)
const inputText = ref('')
const loadingMore = ref(false)
const noMore = ref(false)
const showMembers = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const messages = computed(() => store.messagesCache[props.group.id] ?? [])
const members = computed(() => store.membersCache[props.group.id] ?? [])
const isOwner = computed(() => props.group.owner_id === auth.user?.id)

function scrollToBottom(smooth = false) {
  const el = msgListRef.value
  if (!el) return
  requestAnimationFrame(() => {
    el.scrollTo({ top: el.scrollHeight, behavior: smooth ? 'smooth' : 'auto' })
  })
}

async function loadMessages() {
  noMore.value = false
  const msgs = await store.fetchMessages(props.group.id)
  if (msgs.length < 30) noMore.value = true
  scrollToBottom()
}

async function loadMore() {
  if (loadingMore.value || noMore.value || messages.value.length === 0) return
  loadingMore.value = true
  const firstId = messages.value[0].id
  const el = msgListRef.value
  const prevScrollHeight = el?.scrollHeight ?? 0
  try {
    const older = await store.fetchMessages(props.group.id, firstId)
    if (older.length < 30) noMore.value = true
    requestAnimationFrame(() => {
      if (el) el.scrollTop = el.scrollHeight - prevScrollHeight
    })
  } finally {
    loadingMore.value = false
  }
}

function onScroll() {
  const el = msgListRef.value
  if (el && el.scrollTop < 60) loadMore()
}

function sendText() {
  const text = inputText.value.trim()
  if (!text) return
  ws.send({ type: 'chat_group', to_id: props.group.id, msg_type: 'text', content: { text } })
  inputText.value = ''
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendText()
  }
}

async function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  ;(e.target as HTMLInputElement).value = ''
  const formData = new FormData()
  formData.append('file', file)
  try {
    const res = await fetch('/api/messages/upload', {
      method: 'POST',
      headers: { Authorization: `Bearer ${auth.token}` },
      body: formData,
    })
    const json = await res.json()
    const isImage = file.type.startsWith('image/')
    ws.send({
      type: 'chat_group',
      to_id: props.group.id,
      msg_type: isImage ? 'image' : 'file',
      content: isImage
        ? { url: json.data.url }
        : { url: json.data.url, name: file.name, size: file.size },
    })
  } catch { /* ignore */ }
}

async function kickMember(member: GroupMember) {
  if (!confirm(`确定踢出 ${member.user?.nickname}？`)) return
  await groupApi.kickMember(props.group.id, member.user_id)
  await store.fetchMembers(props.group.id)
}

const { openCard } = useUserCard()

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

watch(() => props.group.id, () => {
  noMore.value = false
  loadMessages()
  store.fetchMembers(props.group.id)
}, { immediate: true })

let prevLen = 0
watchEffect(() => {
  const msgs = store.messagesCache[props.group.id] ?? []
  if (msgs.length > prevLen) {
    const last = msgs[msgs.length - 1]
    const isOwn = Number(last?.from_id) === Number(auth.user?.id)
    const el = msgListRef.value
    const nearBottom = el ? el.scrollHeight - el.scrollTop - el.clientHeight < 120 : true
    if (isOwn || nearBottom) scrollToBottom(true)
  }
  prevLen = msgs.length
})
</script>

<template>
  <div class="group-chat-window">
    <!-- 标题栏 -->
    <header class="chat-header">
      <div class="chat-header-name">{{ group.name }}</div>
      <div class="header-actions">
        <span class="member-count">{{ group.member_count ?? members.length }} 人</span>
        <button class="toggle-members-btn" @click="showMembers = !showMembers" title="成员列表">
          👥
        </button>
      </div>
    </header>

    <div class="chat-body">
      <!-- 消息区 -->
      <div class="msg-list" ref="msgListRef" @scroll="onScroll">
        <div v-if="loadingMore" class="load-tip">加载中…</div>
        <div v-else-if="noMore" class="load-tip">没有更多消息了</div>
        <ChatBubble
          v-for="msg in messages"
          :key="msg.id"
          :msg="msg"
          :showName="true"
          @open-card="openCard"
        />
      </div>

      <!-- 成员面板 -->
      <div v-if="showMembers" class="members-panel">
        <div class="members-title">群成员</div>
        <div class="members-list">
          <div v-for="m in members" :key="m.user_id" class="member-item">
            <Avatar
              :src="getAvatarSrc(m.user?.avatar)"
              :name="m.user?.nickname"
              :size="32"
              style="cursor:pointer;flex-shrink:0"
              @click="openCard(m.user_id)"
            />
            <span class="member-name">
              {{ m.user?.nickname }}
              <span v-if="m.user_id === group.owner_id" class="owner-tag">群主</span>
            </span>
            <button
              v-if="isOwner && m.user_id !== auth.user?.id"
              class="kick-btn"
              @click="kickMember(m)"
            >踢出</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <button class="tool-btn" @click="fileInputRef?.click()" title="发图片/文件">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
          <rect x="3" y="3" width="18" height="18" rx="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <polyline points="21 15 16 10 5 21"/>
        </svg>
      </button>
      <input ref="fileInputRef" type="file" style="display:none" @change="onFileChange" />
    </div>

    <!-- 输入区 -->
    <div class="input-area">
      <textarea
        v-model="inputText"
        class="chat-input"
        placeholder="输入消息，Enter 发送，Shift+Enter 换行"
        rows="3"
        @keydown="onKeydown"
      />
      <button class="send-btn" :disabled="!inputText.trim()" @click="sendText">发送</button>
    </div>
  </div>
</template>
<style scoped>
.group-chat-window {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-surface);
}

.chat-header {
  height: 52px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.chat-header-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.member-count {
  font-size: 12px;
  color: var(--text-tertiary);
}

.toggle-members-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 4px;
  border-radius: 4px;
}

.toggle-members-btn:hover {
  background: var(--bg-hover);
}

.chat-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.msg-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
}

.load-tip {
  text-align: center;
  font-size: 12px;
  color: var(--text-tertiary);
  padding: 8px 0 4px;
}

/* 成员面板 */
.members-panel {
  width: 200px;
  border-left: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
}

.members-title {
  padding: 12px 14px 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.members-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
}

.member-avatar {
  flex-shrink: 0;
}

.member-name {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.owner-tag {
  font-size: 10px;
  color: var(--qq-blue-primary);
  background: var(--qq-blue-light);
  border-radius: 3px;
  padding: 0 4px;
  margin-left: 4px;
}

.kick-btn {
  background: none;
  border: 1px solid var(--border-normal);
  border-radius: 4px;
  font-size: 11px;
  color: var(--color-error);
  padding: 1px 6px;
  cursor: pointer;
  flex-shrink: 0;
}

.kick-btn:hover {
  background: #fff1f0;
}

/* 输入区（复用 ChatWindow 样式） */
.toolbar {
  padding: 6px 16px 0;
  display: flex;
  gap: 4px;
  border-top: 1px solid var(--border-light);
}

.tool-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 4px 6px;
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.tool-btn:hover {
  background: var(--bg-hover);
}

.input-area {
  padding: 8px 16px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.chat-input {
  width: 100%;
  resize: none;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
  font-family: inherit;
  box-sizing: border-box;
}

.send-btn {
  align-self: flex-end;
  background: var(--qq-blue-primary);
  color: white;
  border: none;
  border-radius: 6px;
  padding: 6px 20px;
  font-size: 13px;
  cursor: pointer;
  transition: opacity 0.15s;
}

.send-btn:disabled { opacity: 0.45; cursor: not-allowed; }
.send-btn:not(:disabled):hover { opacity: 0.88; }
</style>
