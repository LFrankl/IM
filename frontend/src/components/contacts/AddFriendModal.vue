<script setup lang="ts">
import { ref } from 'vue'
import Modal from '@/components/common/Modal.vue'
import Avatar from '@/components/common/Avatar.vue'
import { userApi } from '@/api/user'
import { friendApi } from '@/api/friend'
import type { User } from '@/types/user'
import type { ApiResponse } from '@/api/client'

const emit = defineEmits<{ close: [] }>()

const keyword = ref('')
const loading = ref(false)
const results = ref<User[]>([])
const sending = ref<Set<number>>(new Set())
const sent = ref<Set<number>>(new Set())
const error = ref('')

async function search() {
  if (!keyword.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    const res = await userApi.search(keyword.value.trim())
    const body = res.data as unknown as ApiResponse<User[]>
    results.value = body.data ?? []
  } catch {
    error.value = '搜索失败，请重试'
  } finally {
    loading.value = false
  }
}

async function sendRequest(user: User) {
  sending.value.add(user.id)
  try {
    await friendApi.sendRequest(user.id, `你好，我是 ${user.nickname}，请求加你为好友`)
    sent.value.add(user.id)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message
    error.value = msg || '发送失败'
  } finally {
    sending.value.delete(user.id)
  }
}
</script>

<template>
  <Modal title="添加好友" :width="420" @close="emit('close')">
    <div class="search-area">
      <div class="search-input-wrap">
        <input
          v-model="keyword"
          placeholder="搜索用户名或昵称"
          @keyup.enter="search"
          class="search-input"
        />
        <button class="search-btn" :disabled="loading" @click="search">
          {{ loading ? '搜索中...' : '搜索' }}
        </button>
      </div>
      <p v-if="error" class="error-text">{{ error }}</p>
    </div>

    <div v-if="results.length" class="result-list">
      <div v-for="user in results" :key="user.id" class="result-item">
        <Avatar :name="user.nickname" :size="40" :status="user.status" show-status />
        <div class="result-info">
          <div class="result-name">{{ user.nickname }}</div>
          <div class="result-username">{{ user.username }}</div>
        </div>
        <button
          class="add-btn"
          :disabled="sent.has(user.id) || sending.has(user.id)"
          @click="sendRequest(user)"
        >
          {{ sent.has(user.id) ? '已发送' : sending.has(user.id) ? '发送中...' : '加好友' }}
        </button>
      </div>
    </div>

    <div v-else-if="!loading && keyword" class="empty">
      未找到相关用户
    </div>
  </Modal>
</template>

<style scoped>
.search-area { margin-bottom: 16px; }

.search-input-wrap {
  display: flex;
  gap: 8px;
}

.search-input {
  flex: 1;
  height: 36px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-input);
  padding: 0 12px;
  font-size: 13px;
  background: var(--bg-input);
  color: var(--text-primary);
  user-select: text;
}

.search-input:focus {
  border-color: var(--qq-blue-primary);
  box-shadow: 0 0 0 2px var(--qq-blue-light);
}

.search-btn {
  height: 36px;
  padding: 0 16px;
  background: var(--qq-blue-primary);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 13px;
  transition: background var(--transition-fast);
}

.search-btn:hover:not(:disabled) { background: var(--qq-blue-hover); }
.search-btn:disabled { opacity: 0.6; cursor: not-allowed; }

.error-text {
  margin-top: 6px;
  font-size: 12px;
  color: var(--color-error);
}

.result-list { display: flex; flex-direction: column; gap: 4px; }

.result-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: var(--radius-card);
  transition: background var(--transition-fast);
}

.result-item:hover { background: var(--bg-list-item-hover); }

.result-info { flex: 1; min-width: 0; }

.result-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.result-username {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.add-btn {
  height: 30px;
  padding: 0 12px;
  background: var(--qq-blue-primary);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 12px;
  transition: background var(--transition-fast);
  flex-shrink: 0;
}

.add-btn:hover:not(:disabled) { background: var(--qq-blue-hover); }
.add-btn:disabled { background: #ccc; cursor: not-allowed; }

.empty {
  text-align: center;
  padding: 32px 0;
  font-size: 13px;
  color: var(--text-tertiary);
}
</style>
