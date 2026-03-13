<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useGroupStore, type GroupWithMeta } from '@/stores/group'
import { useAuthStore } from '@/stores/auth'
import { groupApi } from '@/api/group'
import type { Message } from '@/types/chat'
import GroupChatWindow from '@/components/chat/GroupChatWindow.vue'
import CreateGroupModal from '@/components/group/CreateGroupModal.vue'
import GroupInviteModal from '@/components/group/GroupInviteModal.vue'
import type { Group } from '@/types/group'
import type { ApiResponse } from '@/api/client'

const store = useGroupStore()
const auth = useAuthStore()

const showCreate = ref(false)
const showInvites = ref(false)
const searchKeyword = ref('')
const searchResults = ref<Group[]>([])

const groups = computed(() => store.myGroups)
const activeGroup = computed(() => store.activeGroup())
const isSearching = computed(() => searchKeyword.value.trim().length > 0)

// 本地实时过滤：我的群组中匹配关键词的
const filteredMyGroups = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return []
  return store.myGroups.filter((g) => g.name.toLowerCase().includes(kw))
})

// 外部搜索结果中不在我群里的
const externalResults = computed(() =>
  searchResults.value.filter((g) => !store.myGroups.find((m) => m.id === g.id))
)

// ── 群消息记录搜索 ──
interface GroupMsgMatch {
  group: GroupWithMeta
  msg: Message
  snippet: string
}
const groupMsgMatches = computed<GroupMsgMatch[]>(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return []
  const results: GroupMsgMatch[] = []
  for (const [gidStr, msgs] of Object.entries(store.messagesCache)) {
    const group = store.myGroups.find((g) => g.id === Number(gidStr))
    if (!group) continue
    for (const msg of [...msgs].reverse()) {
      if (msg.is_recalled || msg.msg_type !== 'text') continue
      const raw = typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content)
      let text = ''
      try { text = (JSON.parse(raw) as { text: string }).text } catch { continue }
      if (!text.toLowerCase().includes(kw)) continue
      const idx = text.toLowerCase().indexOf(kw)
      const start = Math.max(0, idx - 15)
      const end = Math.min(text.length, idx + kw.length + 15)
      const snippet = (start > 0 ? '…' : '') + text.slice(start, end) + (end < text.length ? '…' : '')
      results.push({ group, msg, snippet })
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

function highlight(text: string): string {
  const kw = searchKeyword.value.trim()
  if (!kw) return text
  const re = new RegExp(kw.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
  return text.replace(re, (m) => `<mark>${m}</mark>`)
}

onMounted(() => {
  store.setActiveGroup(null)
  store.fetchMyGroups()
})

onUnmounted(() => {
  store.setActiveGroup(null)
})

function selectGroup(id: number) {
  store.setActiveGroup(id)
}

function selectAndClear(id: number) {
  store.setActiveGroup(id)
  clearSearch()
}

function clearSearch() {
  searchKeyword.value = ''
  searchResults.value = []
}

async function doSearch() {
  const kw = searchKeyword.value.trim()
  if (!kw) { searchResults.value = []; return }
  try {
    const res = await groupApi.search(kw)
    const body = res.data as unknown as ApiResponse<Group[]>
    searchResults.value = body.data ?? []
  } catch { searchResults.value = [] }
}

async function joinGroup(id: number) {
  await groupApi.join(id)
  await store.fetchMyGroups()
  store.setActiveGroup(id)
  clearSearch()
}

function onCreated(g: Group) {
  store.fetchMyGroups()
  store.setActiveGroup(g.id)
}

async function onInviteAccepted() {
  await store.fetchMyGroups()
  store.clearPendingInvites()
  showInvites.value = false
}

async function leaveGroup(g: GroupWithMeta) {
  if (!confirm(`确定退出群组「${g.name}」？`)) return
  await groupApi.leave(g.id)
  store.removeGroup(g.id)
}

async function disbandGroup(g: GroupWithMeta) {
  if (!confirm(`确定解散群组「${g.name}」？此操作不可撤销。`)) return
  await groupApi.disband(g.id)
  store.removeGroup(g.id)
}

function lastMsgPreview(g: GroupWithMeta): string {
  const msg = g.last_message
  if (!msg) return ''
  if (msg.is_recalled) {
    return msg.from_id === auth.user?.id ? '你撤回了一条消息' : `${msg.from?.nickname ?? '对方'} 撤回了一条消息`
  }
  if (msg.msg_type === 'text') {
    try {
      const c = typeof msg.content === 'string' ? JSON.parse(msg.content) : msg.content
      return (c as { text: string }).text
    } catch { return '' }
  }
  if (msg.msg_type === 'image') return '[图片]'
  if (msg.msg_type === 'file') return '[文件]'
  return ''
}

function avatarColor(name: string): string {
  const colors = ['#1677FF', '#52C41A', '#FA8C16', '#EB2F96', '#722ED1', '#13C2C2']
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) % colors.length
  return colors[Math.abs(h)]
}

function getGroupAvatarSrc(avatar: string | undefined): string | null {
  if (!avatar) return null
  if (avatar.startsWith('http')) return avatar
  return `http://localhost:8080${avatar}`
}

function timeStr(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  }
  return `${d.getMonth() + 1}/${d.getDate()}`
}
</script>

<template>
  <div class="view-layout">
    <!-- 左侧群列表 -->
    <aside class="group-list">
      <div class="list-header">
        <span class="list-title">群组</span>
        <div class="header-btns">
          <button class="icon-btn" title="群邀请" @click="showInvites = true">
            <span>📨</span>
            <span v-if="store.pendingInviteCount > 0" class="invite-badge">
              {{ store.pendingInviteCount > 99 ? '99+' : store.pendingInviteCount }}
            </span>
          </button>
          <button class="add-btn" title="创建群组" @click="showCreate = true">＋</button>
        </div>
      </div>

      <div class="search-wrap">
        <div class="search-bar" :class="{ focused: isSearching }">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchKeyword"
            placeholder="搜索群组"
            @input="doSearch"
            @keydown.esc="clearSearch"
          />
          <button v-if="isSearching" class="search-clear" @click="clearSearch">✕</button>
        </div>
      </div>

      <!-- 搜索模式：替换整个列表 -->
      <div v-if="isSearching" class="search-panel">

        <!-- 我的群组匹配 -->
        <template v-if="filteredMyGroups.length > 0">
          <div class="search-section-label">
            我的群组
            <span class="section-count">{{ filteredMyGroups.length }}</span>
          </div>
          <div
            v-for="g in sliced('mygroups', filteredMyGroups)"
            :key="g.id"
            class="group-item"
            :class="{ active: g.id === store.activeGroupId }"
            @click="selectAndClear(g.id)"
          >
            <div class="group-avatar" :style="getGroupAvatarSrc(g.avatar) ? {} : { background: avatarColor(g.name) }">
              <img v-if="getGroupAvatarSrc(g.avatar)" :src="getGroupAvatarSrc(g.avatar)!" class="group-avatar-img" />
              <template v-else>{{ g.name.charAt(0) }}</template>
            </div>
            <div class="group-info">
              <div class="group-row1">
                <span class="group-name" v-html="highlight(g.name)" />
                <span class="group-count">{{ g.member_count ?? 0 }}人</span>
              </div>
              <div class="group-preview">{{ lastMsgPreview(g) }}</div>
            </div>
            <span v-if="g.unread_count && g.unread_count > 0" class="unread-badge">
              {{ g.unread_count > 99 ? '99+' : g.unread_count }}
            </span>
          </div>
          <button v-if="filteredMyGroups.length > COLLAPSE_N" class="section-toggle" @click="toggle('mygroups')">
            <template v-if="!isExpanded('mygroups')">
              查看更多 {{ filteredMyGroups.length - COLLAPSE_N }} 条
              <span class="chevron">›</span>
            </template>
            <template v-else>收起 <span class="chevron up">›</span></template>
          </button>
        </template>

        <!-- 群聊记录 -->
        <template v-if="groupMsgMatches.length > 0">
          <div class="search-section-label">
            群聊记录
            <span class="section-count">{{ groupMsgMatches.length }}</span>
          </div>
          <div
            v-for="item in sliced('groupmsgs', groupMsgMatches)"
            :key="item.msg.id"
            class="group-item"
            :class="{ active: item.group.id === store.activeGroupId }"
            @click="selectAndClear(item.group.id)"
          >
            <div class="group-avatar" :style="getGroupAvatarSrc(item.group.avatar) ? {} : { background: avatarColor(item.group.name) }">
              <img v-if="getGroupAvatarSrc(item.group.avatar)" :src="getGroupAvatarSrc(item.group.avatar)!" class="group-avatar-img" />
              <template v-else>{{ item.group.name.charAt(0) }}</template>
            </div>
            <div class="group-info">
              <div class="group-row1">
                <span class="group-name">{{ item.group.name }}</span>
                <span class="group-count">{{ timeStr(item.msg.created_at) }}</span>
              </div>
              <div class="group-preview">
                <span class="msg-sender-prefix">{{ item.msg.from?.nickname ?? '' }}：</span>
                <span v-html="highlight(item.snippet)" />
              </div>
            </div>
          </div>
          <button v-if="groupMsgMatches.length > COLLAPSE_N" class="section-toggle" @click="toggle('groupmsgs')">
            <template v-if="!isExpanded('groupmsgs')">
              查看更多 {{ groupMsgMatches.length - COLLAPSE_N }} 条
              <span class="chevron">›</span>
            </template>
            <template v-else>收起 <span class="chevron up">›</span></template>
          </button>
        </template>

        <!-- 外部搜索结果 -->
        <template v-if="externalResults.length > 0">
          <div class="search-section-label">
            其他群组
            <span class="section-count">{{ externalResults.length }}</span>
          </div>
          <div
            v-for="g in sliced('external', externalResults)"
            :key="g.id"
            class="group-item"
          >
            <div class="group-avatar" :style="getGroupAvatarSrc(g.avatar) ? {} : { background: avatarColor(g.name) }">
              <img v-if="getGroupAvatarSrc(g.avatar)" :src="getGroupAvatarSrc(g.avatar)!" class="group-avatar-img" />
              <template v-else>{{ g.name.charAt(0) }}</template>
            </div>
            <div class="group-info">
              <span class="group-name">{{ g.name }}</span>
              <div class="group-preview">{{ g.member_count ?? 0 }}人</div>
            </div>
            <button class="join-btn" @click.stop="joinGroup(g.id)">加入</button>
          </div>
          <button v-if="externalResults.length > COLLAPSE_N" class="section-toggle" @click="toggle('external')">
            <template v-if="!isExpanded('external')">
              查看更多 {{ externalResults.length - COLLAPSE_N }} 条
              <span class="chevron">›</span>
            </template>
            <template v-else>收起 <span class="chevron up">›</span></template>
          </button>
        </template>

        <!-- 无结果 -->
        <div v-if="filteredMyGroups.length === 0 && groupMsgMatches.length === 0 && externalResults.length === 0" class="search-empty">
          未找到相关群组
        </div>
      </div>

      <!-- 正常模式：我的群列表 -->
      <div v-else class="group-items">
        <div
          v-for="g in groups"
          :key="g.id"
          class="group-item"
          :class="{ active: g.id === store.activeGroupId }"
          @click="selectGroup(g.id)"
        >
          <div class="group-avatar" :style="getGroupAvatarSrc(g.avatar) ? {} : { background: avatarColor(g.name) }">
            <img v-if="getGroupAvatarSrc(g.avatar)" :src="getGroupAvatarSrc(g.avatar)!" class="group-avatar-img" />
            <template v-else>{{ g.name.charAt(0) }}</template>
          </div>
          <div class="group-info">
            <div class="group-row1">
              <span class="group-name">{{ g.name }}</span>
              <span class="group-count">{{ g.member_count ?? 0 }}人</span>
            </div>
            <div class="group-preview">{{ lastMsgPreview(g) }}</div>
          </div>
          <span v-if="g.unread_count && g.unread_count > 0" class="unread-badge">
            {{ g.unread_count > 99 ? '99+' : g.unread_count }}
          </span>
        </div>
        <div v-if="groups.length === 0" class="empty-groups">
          暂无群组，点击 ＋ 创建
        </div>
      </div>
    </aside>

    <!-- 右侧 -->
    <main class="chat-area">
      <template v-if="activeGroup">
        <GroupChatWindow :group="activeGroup" :key="activeGroup.id" style="flex:1;min-height:0;" />
        <div class="group-actions-bar">
          <button
            v-if="activeGroup.owner_id === auth.user?.id"
            class="danger-btn"
            @click="disbandGroup(activeGroup)"
          >解散群组</button>
          <button v-else class="secondary-btn" @click="leaveGroup(activeGroup)">退出群组</button>
        </div>
      </template>

      <div v-else class="no-chat">
        <div class="no-chat-icon">👨‍👩‍👧‍👦</div>
        <div class="no-chat-text">选择或创建一个群组</div>
      </div>
    </main>
  </div>

  <CreateGroupModal v-if="showCreate" @close="showCreate = false" @created="onCreated" />
  <GroupInviteModal
    v-if="showInvites"
    @close="showInvites = false"
    @accepted="onInviteAccepted"
  />
</template>

<style scoped>
.view-layout {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.group-list {
  width: var(--list-width);
  flex-shrink: 0;
  background: var(--bg-list);
  border-right: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-header {
  height: 52px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.list-title { font-size: 18px; font-weight: 700; color: var(--text-primary); }

.header-btns {
  display: flex;
  align-items: center;
  gap: 4px;
}

.icon-btn {
  position: relative;
  width: 28px; height: 28px;
  border-radius: 6px; font-size: 16px;
  color: var(--text-secondary);
  background: none; border: none;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer;
}
.icon-btn:hover { background: var(--bg-hover); }

.invite-badge {
  position: absolute;
  top: -3px; right: -3px;
  min-width: 15px; height: 15px;
  background: var(--color-badge);
  color: white; font-size: 9px;
  border-radius: 8px; padding: 0 3px;
  display: flex; align-items: center; justify-content: center;
  border: 1.5px solid var(--bg-list);
}

.add-btn {
  width: 28px; height: 28px;
  border-radius: 6px; font-size: 18px;
  color: var(--text-secondary);
  display: flex; align-items: center; justify-content: center;
}
.add-btn:hover { background: var(--bg-hover); color: var(--qq-blue-primary); }

.search-wrap { padding: 0 10px 8px; flex-shrink: 0; }

.search-bar {
  height: 32px; background: #E8E8E8;
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
  display: flex; align-items: center; gap: 4px;
  font-size: 11px; color: var(--text-tertiary);
  padding: 8px 14px 4px; user-select: none;
}
.section-count { font-size: 11px; color: var(--text-tertiary); }
.search-empty {
  text-align: center; color: var(--text-tertiary); font-size: 13px; padding: 40px 0;
}

.section-toggle {
  width: 100%; padding: 6px 14px;
  text-align: left; font-size: 12px;
  color: var(--text-secondary); background: none; border: none;
  cursor: pointer; display: flex; align-items: center; gap: 4px;
  transition: color 0.12s;
}
.section-toggle:hover { color: var(--qq-blue-primary); }

.chevron {
  display: inline-block;
  transform: rotate(90deg);
  transition: transform 0.2s;
  line-height: 1;
}
.chevron.up { transform: rotate(-90deg); }

.msg-sender-prefix { color: var(--text-tertiary); }

.search-panel :deep(mark) {
  background: transparent;
  color: var(--qq-blue-primary);
  font-weight: 600;
}

.group-items { flex: 1; overflow-y: auto; }

.group-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; cursor: pointer; transition: background 0.12s;
}
.group-item:hover { background: var(--bg-hover); }
.group-item.active { background: var(--bg-active); }

.group-avatar {
  width: 40px; height: 40px;
  border-radius: var(--radius-avatar);
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 16px; font-weight: 700; flex-shrink: 0;
  overflow: hidden;
}

.group-avatar-img {
  width: 100%; height: 100%;
  object-fit: cover;
  display: block;
}

.group-info { flex: 1; min-width: 0; }

.group-row1 {
  display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 3px;
}

.group-name {
  font-size: 14px; font-weight: 500; color: var(--text-primary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 140px;
}
.group-count { font-size: 11px; color: var(--text-tertiary); flex-shrink: 0; }
.group-preview { font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.join-btn {
  background: var(--qq-blue-primary); color: white; border: none;
  border-radius: 4px; font-size: 12px; padding: 3px 10px; cursor: pointer; flex-shrink: 0;
}
.join-btn:hover { opacity: 0.88; }

.joined-tag { font-size: 12px; color: var(--text-tertiary); flex-shrink: 0; }

.unread-badge {
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

.empty-groups { text-align: center; color: var(--text-tertiary); font-size: 13px; padding: 40px 0; }

.chat-area { flex: 1; min-width: 0; overflow: hidden; display: flex; flex-direction: column; }

.group-actions-bar {
  padding: 6px 16px; border-top: 1px solid var(--border-light);
  display: flex; justify-content: flex-end; flex-shrink: 0;
}

.danger-btn {
  font-size: 12px; padding: 4px 12px; border-radius: 4px; cursor: pointer;
  color: var(--color-error); border: 1px solid var(--color-error); background: none;
}
.danger-btn:hover { background: #fff1f0; }

.secondary-btn {
  font-size: 12px; padding: 4px 12px; border-radius: 4px; cursor: pointer;
  color: var(--text-secondary); border: 1px solid var(--border-normal); background: none;
}
.secondary-btn:hover { background: var(--bg-hover); }

.no-chat { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; }
.no-chat-icon { font-size: 52px; opacity: 0.2; }
.no-chat-text { font-size: 14px; color: var(--text-tertiary); }
</style>
