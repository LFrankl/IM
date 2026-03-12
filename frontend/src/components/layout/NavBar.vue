<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useContactsStore } from '@/stores/contacts'
import { useChatStore } from '@/stores/chat'
import { useGroupStore } from '@/stores/group'
import { computed, onMounted, ref } from 'vue'
import Avatar from '@/components/common/Avatar.vue'
import ProfileModal from '@/components/common/ProfileModal.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const contacts = useContactsStore()
const chat = useChatStore()
const group = useGroupStore()

const showProfile = ref(false)

const navItems = [
  { name: 'chat',     label: '消息',   icon: '💬' },
  { name: 'contacts', label: '联系人', icon: '👥' },
  { name: 'groups',   label: '群组',   icon: '👨‍👩‍👧‍👦' },
  { name: 'space',    label: 'QQ空间', icon: '✨' },
]

function isActive(name: string) {
  return route.name === name || route.path.startsWith(`/${name}`)
}

function navigate(name: string) {
  router.push({ name })
}

onMounted(async () => {
  if (auth.isLoggedIn) {
    contacts.fetchPendingCount()
  }
})

const chatUnread = computed(() => chat.totalUnread)
const contactsBadge = computed(() => contacts.pendingCount)
const groupUnread = computed(() => group.totalUnread)

function badgeFor(name: string) {
  if (name === 'contacts') return contactsBadge.value
  if (name === 'chat') return chatUnread.value
  if (name === 'groups') return groupUnread.value
  return 0
}

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}
</script>

<template>
  <nav class="nav-bar">
    <!-- 用户头像 -->
    <div class="user-avatar-wrap" @click="showProfile = true">
      <Avatar
        :src="getAvatarSrc(auth.user?.avatar)"
        :name="auth.user?.nickname"
        :size="40"
        :status="auth.user?.status"
        show-status
      />
    </div>

    <!-- 导航入口 -->
    <ul class="nav-items">
      <li
        v-for="item in navItems"
        :key="item.name"
        class="nav-item"
        :class="{ active: isActive(item.name) }"
        @click="navigate(item.name)"
      >
        <div class="nav-icon-wrap">
          <span class="nav-icon">{{ item.icon }}</span>
          <span v-if="badgeFor(item.name) > 0" class="nav-badge">
            {{ badgeFor(item.name) > 99 ? '99+' : badgeFor(item.name) }}
          </span>
        </div>
        <span class="nav-label">{{ item.label }}</span>
      </li>
    </ul>

    <!-- 底部设置 -->
    <div class="nav-bottom">
      <div class="nav-item" title="退出登录" @click="auth.logout().then(() => router.push('/login'))">
        <div class="nav-icon-wrap">
          <span class="nav-icon">⚙️</span>
        </div>
        <span class="nav-label">设置</span>
      </div>
    </div>
  </nav>

  <!-- 资料编辑弹窗 -->
  <ProfileModal v-if="showProfile" @close="showProfile = false" />
</template>

<style scoped>
.nav-bar {
  width: var(--nav-width);
  min-width: var(--nav-width);
  height: 100vh;
  background: var(--bg-nav);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
}

.user-avatar-wrap {
  cursor: pointer;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.nav-items {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  width: 100%;
}

.nav-item {
  width: 52px;
  padding: 6px 4px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  transition: background var(--transition-fast);
}

.nav-item:hover { background: var(--bg-nav-hover); }
.nav-item.active { background: var(--bg-nav-active); }

.nav-icon-wrap {
  position: relative;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-icon { font-size: 18px; line-height: 1; }

.nav-badge {
  position: absolute;
  top: -4px;
  right: -6px;
  min-width: 16px;
  height: 16px;
  background: var(--color-badge);
  color: white;
  font-size: 10px;
  border-radius: 8px;
  padding: 0 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid var(--bg-nav);
}

.nav-label {
  font-size: 10px;
  color: var(--text-nav);
  white-space: nowrap;
}

.nav-item.active .nav-label { color: white; }

.nav-bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>
