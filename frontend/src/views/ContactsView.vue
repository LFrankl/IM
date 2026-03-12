<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Avatar from '@/components/common/Avatar.vue'
import AddFriendModal from '@/components/contacts/AddFriendModal.vue'
import UserProfileModal from '@/components/contacts/UserProfileModal.vue'
import { useContactsStore } from '@/stores/contacts'
import { useChatStore } from '@/stores/chat'
import { friendApi } from '@/api/friend'
import type { Friendship, FriendRequest } from '@/types/user'

const contacts = useContactsStore()
const chat = useChatStore()
const router = useRouter()

const searchKeyword = ref('')
const showAddFriend = ref(false)
const activePanel = ref<'friends' | 'requests'>('friends')
const selectedFriendship = ref<Friendship | null>(null)
const groupCollapsed = ref<Record<string, boolean>>({})

onMounted(async () => {
  await Promise.all([contacts.fetchFriends(), contacts.fetchRequests()])
})

// 搜索过滤
const filteredGroups = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return contacts.friendGroups
  return contacts.friendGroups.map((g) => ({
    ...g,
    friends: g.friends.filter((f) => {
      const name = (f.remark || f.friend?.nickname || '').toLowerCase()
      const username = (f.friend?.username || '').toLowerCase()
      return name.includes(kw) || username.includes(kw)
    }),
  })).filter((g) => g.friends.length > 0)
})

function toggleGroup(groupName: string) {
  groupCollapsed.value[groupName] = !groupCollapsed.value[groupName]
}

function isCollapsed(groupName: string) {
  return !!groupCollapsed.value[groupName]
}

function selectFriend(f: Friendship) {
  selectedFriendship.value = f
  activePanel.value = 'friends'
}

// 好友申请处理
const handlingId = ref<number | null>(null)

async function handleRequest(req: FriendRequest, action: 'accept' | 'reject') {
  handlingId.value = req.id
  try {
    await friendApi.handleRequest(req.id, action)
    contacts.removePendingRequest(req.id)
    if (action === 'accept') {
      // 刷新好友列表
      await contacts.fetchFriends()
    }
  } finally {
    handlingId.value = null
  }
}

function getDisplayName(f: Friendship) {
  return f.remark || f.friend?.nickname || '未知'
}

function sendMessage(f: Friendship) {
  const convId = `private:${f.friend_id}`
  chat.setActiveConv(convId)
  router.push({ name: 'chat', state: { openConv: convId } })
}
</script>

<template>
  <div class="contacts-view">
    <!-- 左侧列表面板 -->
    <aside class="list-panel">
      <!-- 顶部 -->
      <div class="panel-header">
        <h2 class="panel-title">联系人</h2>
        <button class="add-btn" title="添加好友" @click="showAddFriend = true">＋</button>
      </div>

      <!-- 搜索框 -->
      <div class="search-wrap">
        <div class="search-bar">
          <span class="search-icon">🔍</span>
          <input v-model="searchKeyword" placeholder="搜索联系人" />
        </div>
      </div>

      <!-- 好友申请入口 -->
      <div
        class="requests-entry"
        :class="{ active: activePanel === 'requests' }"
        @click="activePanel = 'requests'; selectedFriendship = null"
      >
        <div class="req-icon">📩</div>
        <span class="req-label">新的朋友</span>
        <span v-if="contacts.pendingCount > 0" class="badge">{{ contacts.pendingCount }}</span>
      </div>

      <!-- 好友分组列表 -->
      <div class="friend-list">
        <div v-for="group in filteredGroups" :key="group.name" class="friend-group">
          <!-- 分组标题 -->
          <div class="group-header" @click="toggleGroup(group.name)">
            <span class="group-arrow" :class="{ collapsed: isCollapsed(group.name) }">▶</span>
            <span class="group-name">{{ group.name }}</span>
            <span class="group-count">{{ group.friends.length }}</span>
          </div>

          <!-- 好友列表 -->
          <div v-if="!isCollapsed(group.name)" class="group-friends">
            <div
              v-for="f in group.friends"
              :key="f.friend_id"
              class="friend-item"
              :class="{ active: selectedFriendship?.friend_id === f.friend_id }"
              @click="selectFriend(f)"
            >
              <Avatar
                :name="f.friend?.nickname"
                :size="36"
                :status="f.friend?.status"
                show-status
              />
              <div class="friend-info">
                <span class="friend-name">{{ getDisplayName(f) }}</span>
                <span v-if="f.remark" class="friend-nickname">{{ f.friend?.nickname }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="filteredGroups.length === 0 && !searchKeyword" class="empty-friends">
          暂无好友，点击 ＋ 添加
        </div>
        <div v-else-if="filteredGroups.length === 0" class="empty-friends">
          未找到「{{ searchKeyword }}」
        </div>
      </div>
    </aside>

    <!-- 右侧内容区 -->
    <main class="content-area">
      <!-- 好友申请列表 -->
      <template v-if="activePanel === 'requests'">
        <div class="content-header">
          <h3>新的朋友</h3>
        </div>
        <div class="request-list">
          <div
            v-for="req in contacts.pendingRequests"
            :key="req.id"
            class="request-item"
          >
            <Avatar :name="req.from?.nickname" :size="44" />
            <div class="request-info">
              <div class="request-name">{{ req.from?.nickname }}</div>
              <div class="request-msg">{{ req.message || '请求加你为好友' }}</div>
            </div>
            <div class="request-actions">
              <button
                class="accept-btn"
                :disabled="handlingId === req.id"
                @click="handleRequest(req, 'accept')"
              >同意</button>
              <button
                class="reject-btn"
                :disabled="handlingId === req.id"
                @click="handleRequest(req, 'reject')"
              >拒绝</button>
            </div>
          </div>
          <div v-if="contacts.pendingRequests.length === 0" class="empty-requests">
            暂无新的好友申请
          </div>
        </div>
      </template>

      <!-- 好友资料 -->
      <template v-else-if="selectedFriendship">
        <div class="profile-panel">
          <div class="profile-avatar-area">
            <Avatar
              :name="selectedFriendship.friend?.nickname"
              :size="80"
              :status="selectedFriendship.friend?.status"
              show-status
            />
          </div>
          <div class="profile-nickname">
            {{ selectedFriendship.remark || selectedFriendship.friend?.nickname }}
          </div>
          <div v-if="selectedFriendship.remark" class="profile-real-name">
            {{ selectedFriendship.friend?.nickname }}
          </div>
          <div class="profile-username">{{ selectedFriendship.friend?.username }}</div>
          <div v-if="selectedFriendship.friend?.bio" class="profile-bio">
            {{ selectedFriendship.friend.bio }}
          </div>
          <div class="profile-status-tag" :class="selectedFriendship.friend?.status">
            {{ selectedFriendship.friend?.status === 'online' ? '在线'
               : selectedFriendship.friend?.status === 'busy' ? '忙碌' : '离线' }}
          </div>
          <div class="profile-actions">
            <button class="send-msg-btn" @click="sendMessage(selectedFriendship)">发消息</button>
          </div>
        </div>
      </template>

      <!-- 空状态 -->
      <template v-else>
        <div class="empty-state">
          <div class="empty-icon">👥</div>
          <div class="empty-text">选择联系人查看资料</div>
        </div>
      </template>
    </main>
  </div>

  <!-- 弹窗 -->
  <AddFriendModal v-if="showAddFriend" @close="showAddFriend = false" />
  <UserProfileModal
    v-if="false"
    :friendship="selectedFriendship!"
    @close="selectedFriendship = null"
    @deleted="selectedFriendship = null"
  />
</template>

<style scoped>
.contacts-view {
  display: flex;
  height: 100%;
  overflow: hidden;
}

/* 列表面板 */
.list-panel {
  width: var(--list-width);
  min-width: var(--list-width);
  height: 100%;
  background: var(--bg-list);
  border-right: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  height: 52px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.add-btn {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  font-size: 18px;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.add-btn:hover {
  background: var(--bg-list-item-hover);
  color: var(--qq-blue-primary);
}

.search-wrap {
  padding: 0 10px 8px;
  flex-shrink: 0;
}

.search-bar {
  height: 32px;
  background: #E8E8E8;
  border-radius: var(--radius-search);
  display: flex;
  align-items: center;
  padding: 0 10px;
  gap: 6px;
}

.search-bar:focus-within {
  background: white;
  box-shadow: 0 0 0 1px var(--qq-blue-primary);
}

.search-icon { font-size: 13px; opacity: 0.5; }

.search-bar input {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  background: transparent;
  user-select: text;
}

.search-bar input::placeholder { color: var(--text-tertiary); }

/* 好友申请入口 */
.requests-entry {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  cursor: pointer;
  transition: background var(--transition-fast);
  flex-shrink: 0;
}

.requests-entry:hover { background: var(--bg-list-item-hover); }
.requests-entry.active { background: var(--bg-list-item-active); }

.req-icon { font-size: 20px; width: 36px; text-align: center; }

.req-label {
  flex: 1;
  font-size: 14px;
  color: var(--text-primary);
}

.badge {
  min-width: 18px;
  height: 18px;
  background: var(--color-badge);
  color: white;
  font-size: 11px;
  border-radius: 9px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 好友列表 */
.friend-list {
  flex: 1;
  overflow-y: auto;
  padding-bottom: 8px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  cursor: pointer;
  user-select: none;
}

.group-arrow {
  font-size: 10px;
  color: var(--text-secondary);
  transition: transform var(--transition-fast);
  display: inline-block;
  transform: rotate(90deg);
}

.group-arrow.collapsed { transform: rotate(0deg); }

.group-name {
  flex: 1;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.group-count {
  font-size: 11px;
  color: var(--text-tertiary);
}

.friend-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.friend-item:hover { background: var(--bg-list-item-hover); }
.friend-item.active {
  background: var(--bg-list-item-active);
  border-left: 3px solid var(--qq-blue-primary);
  padding-left: 11px;
}

.friend-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.friend-name {
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.friend-nickname {
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.empty-friends {
  padding: 32px 0;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
}

/* 右侧内容区 */
.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg-surface);
}

.content-header {
  height: 52px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.content-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

/* 好友申请 */
.request-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px 24px;
}

.request-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid var(--border-light);
}

.request-info { flex: 1; min-width: 0; }

.request-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.request-msg {
  font-size: 12px;
  color: var(--text-secondary);
}

.request-actions { display: flex; gap: 8px; }

.accept-btn {
  height: 30px;
  padding: 0 14px;
  background: var(--qq-blue-primary);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 13px;
  transition: background var(--transition-fast);
}

.accept-btn:hover:not(:disabled) { background: var(--qq-blue-hover); }
.accept-btn:disabled { opacity: 0.6; cursor: not-allowed; }

.reject-btn {
  height: 30px;
  padding: 0 14px;
  border: 1px solid var(--border-normal);
  border-radius: var(--radius-btn);
  font-size: 13px;
  color: var(--text-secondary);
  transition: background var(--transition-fast);
}

.reject-btn:hover:not(:disabled) { background: var(--bg-list-item-hover); }
.reject-btn:disabled { opacity: 0.6; cursor: not-allowed; }

.empty-requests {
  padding: 48px 0;
  text-align: center;
  font-size: 14px;
  color: var(--text-tertiary);
}

/* 好友资料面板 */
.profile-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 24px 24px;
}

.profile-avatar-area { margin-bottom: 16px; }

.profile-nickname {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.profile-real-name {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.profile-username {
  font-size: 13px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}

.profile-bio {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  max-width: 300px;
  text-align: center;
  line-height: 1.6;
}

.profile-status-tag {
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 10px;
  margin-bottom: 24px;
}

.profile-status-tag.online { background: #f6ffed; color: var(--color-success); }
.profile-status-tag.busy   { background: #fffbe6; color: var(--color-warning); }
.profile-status-tag.offline { background: #f5f5f5; color: var(--color-offline); }

.profile-actions { display: flex; gap: 12px; }

.send-msg-btn {
  height: 38px;
  padding: 0 24px;
  background: var(--qq-blue-primary);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 14px;
}

.send-msg-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.empty-icon { font-size: 48px; opacity: 0.3; }

.empty-text {
  font-size: 14px;
  color: var(--text-tertiary);
}
</style>
