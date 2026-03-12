<script setup lang="ts">
import { ref, computed, watch, watchEffect, nextTick, onMounted, onUnmounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { useWS } from '@/composables/useWS'
import { useUserCard } from '@/composables/useUserCard'
import { chatApi } from '@/api/chat'
import ChatBubble from './ChatBubble.vue'
import EmojiPicker from '@/components/common/EmojiPicker.vue'
import type { Conversation } from '@/types/chat'

const props = defineProps<{ conv: Conversation }>()

const chat = useChatStore()
const auth = useAuthStore()
const ws = useWS()

const msgListRef = ref<HTMLElement | null>(null)
const inputText = ref('')
const loadingMore = ref(false)
const noMore = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const showEmoji = ref(false)
const emojiWrapRef = ref<HTMLElement | null>(null)
const { openCard } = useUserCard()

function insertEmoji(emoji: string) {
  const el = textareaRef.value
  if (!el) { inputText.value += emoji; return }
  const start = el.selectionStart ?? inputText.value.length
  const end = el.selectionEnd ?? inputText.value.length
  inputText.value = inputText.value.slice(0, start) + emoji + inputText.value.slice(end)
  nextTick(() => {
    el.focus()
    const pos = start + [...emoji].length
    el.setSelectionRange(pos, pos)
  })
}

function onDocMousedown(e: MouseEvent) {
  if (showEmoji.value && emojiWrapRef.value && !emojiWrapRef.value.contains(e.target as Node)) {
    showEmoji.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', onDocMousedown))
onUnmounted(() => document.removeEventListener('mousedown', onDocMousedown))

const messages = computed(() => chat.messagesCache[props.conv.id] ?? [])

function scrollToBottom(smooth = false) {
  nextTick(() => {
    const el = msgListRef.value
    if (!el) return
    el.scrollTo({ top: el.scrollHeight, behavior: smooth ? 'smooth' : 'auto' })
  })
}

// 初次加载消息
async function loadMessages() {
  noMore.value = false
  const msgs = await chat.fetchMessages(props.conv.target_id)
  if (msgs.length < 30) noMore.value = true
  scrollToBottom()
  // 标已读
  chat.clearUnread(props.conv.id)
  if (props.conv.unread_count > 0) {
    chatApi.markRead(props.conv.target_id)
  }
}

// 上翻加载更多
async function loadMore() {
  if (loadingMore.value || noMore.value || messages.value.length === 0) return
  loadingMore.value = true
  const firstId = messages.value[0].id
  const el = msgListRef.value
  const prevScrollHeight = el?.scrollHeight ?? 0

  try {
    const older = await chat.fetchMessages(props.conv.target_id, firstId)
    if (older.length < 30) noMore.value = true
    // 保持滚动位置
    nextTick(() => {
      if (el) el.scrollTop = el.scrollHeight - prevScrollHeight
    })
  } finally {
    loadingMore.value = false
  }
}

function onScroll() {
  const el = msgListRef.value
  if (!el) return
  if (el.scrollTop < 60) loadMore()
}

// 发送文字
function sendText() {
  const text = inputText.value.trim()
  if (!text) return
  ws.send({
    type: 'chat_private',
    to_id: props.conv.target_id,
    msg_type: 'text',
    content: { text },
  })
  inputText.value = ''
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendText()
  }
}

// 图片/文件上传
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

    if (isImage) {
      ws.send({
        type: 'chat_private',
        to_id: props.conv.target_id,
        msg_type: 'image',
        content: { url: json.data.url },
      })
    } else {
      ws.send({
        type: 'chat_private',
        to_id: props.conv.target_id,
        msg_type: 'file',
        content: { url: json.data.url, name: file.name, size: file.size },
      })
    }
  } catch {
    // 上传失败静默
  }
}

// 切换会话时重新加载
watch(() => props.conv.id, () => {
  noMore.value = false
  loadMessages()
}, { immediate: true })

// 新消息自动滚动到底
let prevLen = 0
watchEffect(() => {
  const msgs = chat.messagesCache[props.conv.id] ?? []
  if (msgs.length > prevLen) {
    const last = msgs[msgs.length - 1]
    const isOwn = last?.from_id === auth.user?.id
    const el = msgListRef.value
    const nearBottom = el ? el.scrollHeight - el.scrollTop - el.clientHeight < 120 : true
    if (isOwn || nearBottom) scrollToBottom(true)
  }
  prevLen = msgs.length
})
</script>

<template>
  <div class="chat-window">
    <!-- 顶部标题栏 -->
    <header class="chat-header">
      <div class="chat-header-name">{{ conv.name }}</div>
    </header>

    <!-- 消息列表 -->
    <div class="msg-list" ref="msgListRef" @scroll="onScroll">
      <div v-if="loadingMore" class="load-more-tip">加载中…</div>
      <div v-else-if="noMore" class="load-more-tip">没有更多消息了</div>

      <ChatBubble
        v-for="msg in messages"
        :key="msg.id"
        :msg="msg"
        :showName="false"
        @open-card="openCard"
      />
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="emoji-wrap" ref="emojiWrapRef">
        <button class="tool-btn" title="表情" @click="showEmoji = !showEmoji">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="12" cy="12" r="10"/>
            <path d="M8 14s1.5 2 4 2 4-2 4-2"/>
            <line x1="9" y1="9" x2="9.01" y2="9" stroke-width="2.5" stroke-linecap="round"/>
            <line x1="15" y1="9" x2="15.01" y2="9" stroke-width="2.5" stroke-linecap="round"/>
          </svg>
        </button>
        <EmojiPicker v-if="showEmoji" @pick="insertEmoji" />
      </div>
      <button class="tool-btn" title="发送图片/文件" @click="fileInputRef?.click()">
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
        ref="textareaRef"
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
.chat-window {
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
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.chat-header-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.msg-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
}

.load-more-tip {
  text-align: center;
  font-size: 12px;
  color: var(--text-tertiary);
  padding: 8px 0 4px;
}

.emoji-wrap {
  position: relative;
}

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
  color: var(--text-primary);
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

.send-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.send-btn:not(:disabled):hover {
  opacity: 0.88;
}
</style>
