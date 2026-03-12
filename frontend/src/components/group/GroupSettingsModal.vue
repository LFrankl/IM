<script setup lang="ts">
import { ref, computed } from 'vue'
import { groupApi } from '@/api/group'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/user'
import type { GroupWithMeta } from '@/stores/group'
import type { ApiResponse } from '@/api/client'
import type { User } from '@/types/user'

const props = defineProps<{ group: GroupWithMeta }>()
const emit = defineEmits<{ close: []; updated: [] }>()

const auth = useAuthStore()
const isOwner = computed(() => props.group.owner_id === auth.user?.id)

// 群设置
const allowInvite = ref(props.group.allow_invite ?? true)
const savingSettings = ref(false)

async function saveSettings() {
  savingSettings.value = true
  try {
    await groupApi.updateSettings(props.group.id, allowInvite.value)
    emit('updated')
  } finally {
    savingSettings.value = false
  }
}

// 邀请成员
const searchKeyword = ref('')
const searchResults = ref<User[]>([])
const inviting = ref<number | null>(null)
const invitedSet = ref<Set<number>>(new Set())

async function doSearch() {
  const kw = searchKeyword.value.trim()
  if (!kw) { searchResults.value = []; return }
  try {
    const res = await userApi.search(kw)
    const body = res.data as unknown as ApiResponse<User[]>
    searchResults.value = (body.data ?? []).filter(u => u.id !== auth.user?.id)
  } catch { searchResults.value = [] }
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

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}
</script>

<template>
  <Teleport to="body">
    <div class="mask" @click.self="emit('close')">
      <div class="dialog">
        <div class="dialog-header">
          <span class="dialog-title">群设置</span>
          <button class="close-btn" @click="emit('close')">✕</button>
        </div>

        <!-- 群主：设置区 -->
        <section v-if="isOwner" class="section">
          <div class="section-title">群主设置</div>
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

        <!-- 邀请成员（群主 or allow_invite=true 的成员） -->
        <section v-if="isOwner || group.allow_invite" class="section">
          <div class="section-title">邀请成员</div>
          <div class="search-row">
            <input
              v-model="searchKeyword"
              placeholder="搜索用户昵称"
              class="search-input"
              @input="doSearch"
            />
          </div>
          <div class="search-results">
            <div v-for="u in searchResults" :key="u.id" class="user-row">
              <img
                v-if="getAvatarSrc(u.avatar)"
                :src="getAvatarSrc(u.avatar)"
                class="user-avatar"
              />
              <div v-else class="user-avatar-fallback">{{ u.nickname?.charAt(0) }}</div>
              <span class="user-name">{{ u.nickname }}</span>
              <button
                class="invite-btn"
                :disabled="invitedSet.has(u.id) || inviting === u.id"
                @click="invite(u.id)"
              >
                {{ invitedSet.has(u.id) ? '已邀请' : inviting === u.id ? '邀请中…' : '邀请' }}
              </button>
            </div>
            <div v-if="searchKeyword && searchResults.length === 0" class="empty-tip">
              未找到用户
            </div>
          </div>
        </section>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.mask {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 1500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog {
  background: #fff;
  border-radius: 12px;
  width: 400px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0,0,0,0.15);
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.dialog-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 14px;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}
.close-btn:hover { background: var(--bg-hover); }

.section {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-light);
}
.section:last-child { border-bottom: none; }

.section-title {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 12px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.setting-label {
  font-size: 14px;
  color: var(--text-primary);
}

/* Toggle switch */
.toggle {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
}
.toggle input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  inset: 0;
  background: #ccc;
  border-radius: 22px;
  cursor: pointer;
  transition: background 0.2s;
}
.slider::before {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
}
.toggle input:checked + .slider { background: var(--qq-blue-primary); }
.toggle input:checked + .slider::before { transform: translateX(18px); }

.save-btn {
  background: var(--qq-blue-primary);
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 6px 18px;
  font-size: 13px;
  cursor: pointer;
}
.save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.save-btn:not(:disabled):hover { opacity: 0.88; }

.search-row { margin-bottom: 10px; }

.search-input {
  width: 100%;
  height: 34px;
  padding: 0 10px;
  border: 1px solid var(--border-input);
  border-radius: 6px;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
}
.search-input:focus { border-color: var(--qq-blue-primary); }

.search-results {
  max-height: 220px;
  overflow-y: auto;
}

.user-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 0;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.user-avatar-fallback {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--qq-blue-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.user-name {
  flex: 1;
  font-size: 14px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.invite-btn {
  background: var(--qq-blue-primary);
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  padding: 3px 12px;
  cursor: pointer;
  flex-shrink: 0;
}
.invite-btn:disabled {
  background: var(--bg-hover);
  color: var(--text-tertiary);
  cursor: not-allowed;
}
.invite-btn:not(:disabled):hover { opacity: 0.88; }

.empty-tip {
  font-size: 13px;
  color: var(--text-tertiary);
  text-align: center;
  padding: 16px 0;
}
</style>
