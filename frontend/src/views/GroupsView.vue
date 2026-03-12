<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useGroupStore, type GroupWithMeta } from '@/stores/group'
import { useAuthStore } from '@/stores/auth'
import { groupApi } from '@/api/group'
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

onMounted(() => {
  store.setActiveGroup(null)
  store.fetchMyGroups()
})

function selectGroup(id: number) {
  store.setActiveGroup(id)
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
  searchKeyword.value = ''
  searchResults.value = []
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
        <div class="search-bar">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchKeyword"
            placeholder="搜索群组"
            @input="doSearch"
            @keydown.enter="doSearch"
          />
        </div>
      </div>

      <!-- 搜索结果 -->
      <div v-if="searchKeyword && searchResults.length > 0" class="search-results">
        <div class="search-section-title">搜索结果</div>
        <div v-for="g in searchResults" :key="g.id" class="group-item">
          <div class="group-avatar" :style="{ background: avatarColor(g.name) }">
            {{ g.name.charAt(0) }}
          </div>
          <div class="group-info">
            <span class="group-name">{{ g.name }}</span>
          </div>
          <button
            v-if="!groups.find(myG => myG.id === g.id)"
            class="join-btn"
            @click="joinGroup(g.id)"
          >加入</button>
          <span v-else class="joined-tag">已加入</span>
        </div>
      </div>

      <!-- 我的群 -->
      <div class="group-items">
        <div
          v-for="g in groups"
          :key="g.id"
          class="group-item"
          :class="{ active: g.id === store.activeGroupId }"
          @click="selectGroup(g.id)"
        >
          <div class="group-avatar" :style="{ background: avatarColor(g.name) }">
            {{ g.name.charAt(0) }}
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
        <div v-if="groups.length === 0 && !searchKeyword" class="empty-groups">
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
}
.search-bar:focus-within { background: white; box-shadow: 0 0 0 1px var(--qq-blue-primary); }
.search-icon { font-size: 13px; opacity: 0.5; }
.search-bar input { flex: 1; font-size: 13px; color: var(--text-primary); background: transparent; user-select: text; }

.search-results { border-bottom: 1px solid var(--border-light); padding-bottom: 4px; flex-shrink: 0; }
.search-section-title { font-size: 11px; color: var(--text-tertiary); padding: 6px 14px 2px; }

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
