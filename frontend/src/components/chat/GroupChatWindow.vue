<script setup lang="ts">
import { ref, computed, watch, watchEffect, nextTick, onMounted, onUnmounted } from 'vue'
import { useGroupStore } from '@/stores/group'
import { useAuthStore } from '@/stores/auth'
import { useWS } from '@/composables/useWS'
import { useUserCard } from '@/composables/useUserCard'
import { groupApi } from '@/api/group'
import { userApi } from '@/api/user'
import ChatBubble from '@/components/chat/ChatBubble.vue'
import Avatar from '@/components/common/Avatar.vue'
import EmojiPicker from '@/components/common/EmojiPicker.vue'
import type { GroupWithMeta } from '@/stores/group'
import type { GroupMember } from '@/types/group'
import type { User } from '@/types/user'
import type { ApiResponse } from '@/api/client'

const props = defineProps<{ group: GroupWithMeta }>()

const store = useGroupStore()
const auth = useAuthStore()
const ws = useWS()

const msgListRef = ref<HTMLElement | null>(null)
const inputText = ref('')
const loadingMore = ref(false)
const noMore = ref(false)
const showPanel = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const showEmoji = ref(false)
const emojiWrapRef = ref<HTMLElement | null>(null)

const selStart = ref(0)
const selEnd = ref(0)

function syncSel() {
  const el = textareaRef.value
  if (!el) return
  selStart.value = el.selectionStart ?? 0
  selEnd.value = el.selectionEnd ?? 0
}

// 群设置（面板内）
const allowInvite = ref(props.group.allow_invite ?? true)
const savingSettings = ref(false)

// 群头像上传
const avatarInputRef = ref<HTMLInputElement | null>(null)
const uploadingAvatar = ref(false)

// 邀请成员（面板内）
const inviteKeyword = ref('')
const inviteResults = ref<User[]>([])
const inviting = ref<number | null>(null)
const invitedSet = ref<Set<number>>(new Set())
const showInviteSearch = ref(false)

const MAX_MEMBERS_SHOWN = 20

const messages = computed(() => store.messagesCache[props.group.id] ?? [])
const members = computed(() => store.membersCache[props.group.id] ?? [])
const isOwner = computed(() => props.group.owner_id === auth.user?.id)
const canInvite = computed(() => isOwner.value || props.group.allow_invite)
const shownMembers = computed(() => members.value.slice(0, MAX_MEMBERS_SHOWN))

const groupAvatarSrc = computed(() => {
  const av = props.group.avatar
  if (!av) return undefined
  if (av.startsWith('http')) return av
  return `http://localhost:8080${av}`
})

function avatarBgColor(name: string): string {
  const colors = ['#1677FF', '#52C41A', '#FA8C16', '#EB2F96', '#722ED1', '#13C2C2']
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) % colors.length
  return colors[Math.abs(h)]
}

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

// ---- 滚动 ----
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

// ---- 发送消息 ----
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

function insertEmoji(emoji: string) {
  const el = textareaRef.value
  if (!el) { inputText.value += emoji; return }
  const start = selStart.value
  const end = selEnd.value
  inputText.value = inputText.value.slice(0, start) + emoji + inputText.value.slice(end)
  const newPos = start + emoji.length
  selStart.value = newPos
  selEnd.value = newPos
  nextTick(() => {
    el.focus()
    el.setSelectionRange(newPos, newPos)
  })
}

function onDocMousedown(e: MouseEvent) {
  if (showEmoji.value && emojiWrapRef.value && !emojiWrapRef.value.contains(e.target as Node)) {
    showEmoji.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', onDocMousedown))
onUnmounted(() => document.removeEventListener('mousedown', onDocMousedown))

// ---- 图片/文件上传 ----
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

// ---- 踢人 ----
async function kickMember(member: GroupMember) {
  if (!confirm(`确定踢出 ${member.user?.nickname}？`)) return
  await groupApi.kickMember(props.group.id, member.user_id)
  await store.fetchMembers(props.group.id)
}

// ---- 群头像上传 ----
async function handleAvatarChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  ;(e.target as HTMLInputElement).value = ''
  uploadingAvatar.value = true
  try {
    const res = await groupApi.updateAvatar(props.group.id, file)
    const body = res.data as unknown as ApiResponse<{ avatar: string }>
    if (body.data?.avatar) {
      store.updateGroupAvatar(props.group.id, body.data.avatar)
    }
  } catch (err: any) {
    alert(err?.response?.data?.message ?? '上传失败')
  } finally {
    uploadingAvatar.value = false
  }
}

// ---- 群设置 ----
async function saveSettings() {
  savingSettings.value = true
  try {
    await groupApi.updateSettings(props.group.id, allowInvite.value)
  } finally {
    savingSettings.value = false
  }
}

// ---- 邀请成员 ----
async function doInviteSearch() {
  const kw = inviteKeyword.value.trim()
  if (!kw) { inviteResults.value = []; return }
  try {
    const res = await userApi.search(kw)
    const body = res.data as unknown as ApiResponse<User[]>
    inviteResults.value = (body.data ?? []).filter(u => u.id !== auth.user?.id)
  } catch { inviteResults.value = [] }
}

async function invite(userId: number) {
  inviting.value = userId
  try {
    await groupApi.inviteMember(props.group.id, userId)
    invitedSet.value = new Set([...invitedSet.value, userId])
  } catch (e: any) {
    alert(e?.response?.data?.message ?? '邀请失败')
  } finally {
    inviting.value = null
  }
}

const { openCard } = useUserCard()

watch(() => props.group.id, () => {
  noMore.value = false
  allowInvite.value = props.group.allow_invite ?? true
  inviteKeyword.value = ''
  inviteResults.value = []
  invitedSet.value = new Set()
  showInviteSearch.value = false
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
        <!-- 三横杠：同时控制群设置+成员面板 -->
        <button class="hamburger-btn" :class="{ active: showPanel }" title="群信息" @click="showPanel = !showPanel">
          <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor">
            <rect x="2" y="3" width="14" height="2" rx="1"/>
            <rect x="2" y="8" width="14" height="2" rx="1"/>
            <rect x="2" y="13" width="14" height="2" rx="1"/>
          </svg>
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

      <!-- 右侧面板：群信息（设置 + 邀请 + 成员） -->
      <div v-if="showPanel" class="info-panel">
        <div class="panel-header">
          <span class="panel-title">群信息</span>
          <button class="panel-close" @click="showPanel = false">✕</button>
        </div>

        <div class="panel-body">
          <!-- 群头像 -->
          <div class="panel-avatar-section">
            <div
              class="panel-avatar-wrap"
              :class="{ 'is-owner': isOwner }"
              :title="isOwner ? '点击更换群头像' : ''"
              @click="isOwner && avatarInputRef?.click()"
            >
              <img v-if="groupAvatarSrc" :src="groupAvatarSrc" class="panel-avatar-img" />
              <div
                v-else
                class="panel-avatar-fallback"
                :style="{ background: avatarBgColor(group.name) }"
              >
                {{ group.name.charAt(0) }}
              </div>
              <div v-if="isOwner" class="avatar-edit-overlay">
                <span v-if="uploadingAvatar" class="avatar-uploading">…</span>
                <span v-else>更换</span>
              </div>
            </div>
            <input
              ref="avatarInputRef"
              type="file"
              style="display:none"
              accept="image/jpeg,image/png,image/gif,image/webp"
              @change="handleAvatarChange"
            />
            <div class="panel-group-name">{{ group.name }}</div>
            <div class="panel-member-count">{{ group.member_count ?? members.length }} 名成员</div>
          </div>

          <!-- 群主设置 -->
          <section v-if="isOwner" class="panel-section">
            <div class="panel-section-title">群主设置</div>
            <div class="setting-row">
              <span class="setting-label">允许成员邀请新人</span>
              <label class="toggle">
                <input type="checkbox" v-model="allowInvite" />
                <span class="slider" />
              </label>
            </div>
            <button class="save-btn" :disabled="savingSettings" @click="saveSettings">
              {{ savingSettings ? '保存中…' : '保存' }}
            </button>
          </section>

          <!-- 群成员网格 + 邀请 -->
          <section class="panel-section">
            <div class="panel-section-title">
              群成员（{{ members.length }}）
            </div>

            <div class="members-grid">
              <!-- 成员头像格 -->
              <div
                v-for="m in shownMembers"
                :key="m.user_id"
                class="grid-cell"
                :title="m.user?.nickname"
                @click="openCard(m.user_id)"
              >
                <div class="grid-avatar-wrap">
                  <Avatar
                    :src="getAvatarSrc(m.user?.avatar)"
                    :name="m.user?.nickname"
                    :size="36"
                  />
                  <!-- 群主踢人按钮（hover 出现） -->
                  <button
                    v-if="isOwner && m.user_id !== auth.user?.id"
                    class="kick-x"
                    @click.stop="kickMember(m)"
                    title="踢出"
                  >×</button>
                  <!-- 群主标识 -->
                  <span v-if="m.user_id === group.owner_id" class="crown-badge">♛</span>
                </div>
                <span class="grid-name">{{ m.user?.nickname }}</span>
              </div>

              <!-- + 邀请按钮 -->
              <div
                class="grid-cell"
                :class="canInvite ? 'invite-plus' : 'invite-plus-disabled'"
                :title="canInvite ? '邀请新成员' : '群主未开放邀请权限'"
                @click="canInvite && (showInviteSearch = !showInviteSearch)"
              >
                <div class="grid-plus-wrap">
                  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                    <line x1="9" y1="3" x2="9" y2="15"/>
                    <line x1="3" y1="9" x2="15" y2="9"/>
                  </svg>
                </div>
                <span class="grid-name">邀请</span>
              </div>
            </div>

            <!-- 超出上限提示 -->
            <div v-if="members.length > MAX_MEMBERS_SHOWN" class="members-more">
              仅显示前 {{ MAX_MEMBERS_SHOWN }} 位成员
            </div>

            <!-- 邀请搜索框（点击 + 展开） -->
            <div v-if="showInviteSearch" class="invite-panel">
              <input
                v-model="inviteKeyword"
                placeholder="搜索用户昵称"
                class="invite-search"
                @input="doInviteSearch"
                autofocus
              />
              <div class="invite-results">
                <div v-for="u in inviteResults" :key="u.id" class="invite-row">
                  <Avatar :src="getAvatarSrc(u.avatar)" :name="u.nickname" :size="26" />
                  <span class="invite-name">{{ u.nickname }}</span>
                  <button
                    class="invite-btn"
                    :disabled="invitedSet.has(u.id) || inviting === u.id"
                    @click="invite(u.id)"
                  >
                    {{ invitedSet.has(u.id) ? '已邀请' : inviting === u.id ? '…' : '邀请' }}
                  </button>
                </div>
                <div v-if="inviteKeyword && inviteResults.length === 0" class="invite-empty">
                  未找到用户
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
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
        ref="textareaRef"
        v-model="inputText"
        class="chat-input"
        placeholder="输入消息，Enter 发送，Shift+Enter 换行"
        rows="3"
        @keydown="onKeydown"
        @click="syncSel"
        @keyup="syncSel"
        @select="syncSel"
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
  padding: 0 16px 0 20px;
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

.hamburger-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: none;
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.12s, color 0.12s;
}
.hamburger-btn:hover { background: var(--bg-hover); color: var(--text-primary); }
.hamburger-btn.active { background: var(--bg-active); color: var(--qq-blue-primary); }

.chat-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.msg-list {
  position: relative;
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

/* ---- 右侧面板 ---- */
.info-panel {
  width: 240px;
  border-left: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--bg-surface);
}

.panel-header {
  height: 44px;
  padding: 0 12px 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.panel-close {
  background: none;
  border: none;
  font-size: 13px;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}
.panel-close:hover { background: var(--bg-hover); }

.panel-body {
  flex: 1;
  overflow-y: auto;
}

/* 群头像区域 */
.panel-avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 16px 16px;
  border-bottom: 1px solid var(--border-light);
}

.panel-avatar-wrap {
  position: relative;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-avatar);
  overflow: hidden;
  flex-shrink: 0;
  margin-bottom: 10px;
}

.panel-avatar-wrap.is-owner {
  cursor: pointer;
}

.panel-avatar-wrap.is-owner:hover .avatar-edit-overlay {
  opacity: 1;
}

.panel-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.panel-avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
  font-weight: 700;
}

.avatar-edit-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 12px;
  opacity: 0;
  transition: opacity 0.15s;
}

.avatar-uploading {
  font-size: 18px;
  animation: pulse 0.8s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.panel-group-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  text-align: center;
}

.panel-member-count {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 3px;
}

/* 通用 section */
.panel-section {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-light);
}
.panel-section:last-child { border-bottom: none; }

.panel-section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 10px;
}

/* 设置 */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.setting-label {
  font-size: 13px;
  color: var(--text-primary);
}

.toggle {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
}
.toggle input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  inset: 0;
  background: #ccc;
  border-radius: 20px;
  cursor: pointer;
  transition: background 0.2s;
}
.slider::before {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
}
.toggle input:checked + .slider { background: var(--qq-blue-primary); }
.toggle input:checked + .slider::before { transform: translateX(16px); }

.save-btn {
  background: var(--qq-blue-primary);
  color: #fff;
  border: none;
  border-radius: 5px;
  padding: 5px 14px;
  font-size: 12px;
  cursor: pointer;
}
.save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.save-btn:not(:disabled):hover { opacity: 0.88; }

/* ---- 成员网格 ---- */
.members-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px 2px;
  margin-bottom: 4px;
}

.grid-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 6px 2px 4px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.12s;
  min-width: 0;
}
.grid-cell:hover { background: var(--bg-hover); }

.grid-avatar-wrap {
  position: relative;
  flex-shrink: 0;
}

/* 踢出 × 按钮（群主 hover 时显示） */
.kick-x {
  display: none;
  position: absolute;
  top: -4px;
  right: -4px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: rgba(255, 59, 48, 0.9);
  color: white;
  border: none;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  line-height: 1;
  padding: 0;
  z-index: 1;
}
.grid-cell:hover .kick-x { display: flex; }

/* 群主皇冠角标 */
.crown-badge {
  position: absolute;
  bottom: -3px;
  right: -3px;
  font-size: 10px;
  line-height: 1;
  filter: drop-shadow(0 0 1px rgba(0,0,0,0.3));
}

.grid-name {
  font-size: 10px;
  color: var(--text-secondary);
  text-align: center;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* + 邀请按钮格 */
.grid-plus-wrap {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-avatar);
  border: 1.5px dashed var(--border-normal);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary);
  transition: border-color 0.12s, color 0.12s;
}

.invite-plus { cursor: pointer; }
.invite-plus:hover .grid-plus-wrap {
  border-color: var(--qq-blue-primary);
  color: var(--qq-blue-primary);
}

.invite-plus-disabled { cursor: not-allowed; opacity: 0.5; }
.invite-plus-disabled:hover { background: transparent; }

.members-more {
  font-size: 11px;
  color: var(--text-tertiary);
  text-align: center;
  padding: 2px 0 6px;
}

/* 邀请搜索面板 */
.invite-panel {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border-light);
}

.invite-search {
  width: 100%;
  height: 30px;
  padding: 0 8px;
  border: 1px solid var(--border-input);
  border-radius: 5px;
  font-size: 12px;
  outline: none;
  box-sizing: border-box;
  margin-bottom: 8px;
}
.invite-search:focus { border-color: var(--qq-blue-primary); }

.invite-results { display: flex; flex-direction: column; gap: 4px; }

.invite-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.invite-name {
  flex: 1;
  font-size: 12px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.invite-btn {
  background: var(--qq-blue-primary);
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 11px;
  padding: 2px 8px;
  cursor: pointer;
  flex-shrink: 0;
}
.invite-btn:disabled {
  background: var(--bg-hover);
  color: var(--text-tertiary);
  cursor: not-allowed;
}
.invite-btn:not(:disabled):hover { opacity: 0.88; }

.invite-empty {
  font-size: 12px;
  color: var(--text-tertiary);
  text-align: center;
  padding: 8px 0;
}

/* 输入区 */
.emoji-wrap { position: relative; }

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
.tool-btn:hover { background: var(--bg-hover); }

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
