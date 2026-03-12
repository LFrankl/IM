<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/user'
import { spaceApi, type SpacePost } from '@/api/space'
import Avatar from '@/components/common/Avatar.vue'
import type { User } from '@/types/user'
import type { ApiResponse } from '@/api/client'

const props = defineProps<{ userId: number }>()
const emit = defineEmits<{ close: [] }>()

const router = useRouter()
const auth = useAuthStore()

const user = ref<User | null>(null)
const posts = ref<SpacePost[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [userRes, postsRes] = await Promise.all([
      userApi.getById(props.userId),
      spaceApi.getUserPosts(props.userId),
    ])
    const userBody = userRes.data as unknown as ApiResponse<User>
    const postsBody = postsRes.data as unknown as ApiResponse<SpacePost[]>
    user.value = userBody.data
    posts.value = (postsBody.data ?? []).slice(0, 3)
  } finally {
    loading.value = false
  }
})

function getImgSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

function parseImages(raw: string): string[] {
  try { return JSON.parse(raw) ?? [] } catch { return [] }
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return '刚刚'
  if (m < 60) return `${m}分钟前`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}小时前`
  const d = Math.floor(h / 24)
  return d < 30 ? `${d}天前` : new Date(dateStr).toLocaleDateString()
}

function goToSpace() {
  emit('close')
  if (props.userId === auth.user?.id) {
    router.push({ name: 'space' })
  } else {
    router.push({ name: 'space', params: { userId: props.userId } })
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="card-mask" @click.self="emit('close')">
      <div class="user-card">
        <!-- 封面 -->
        <div
          class="card-cover"
          :style="user?.cover ? { backgroundImage: `url(${getImgSrc(user.cover)})` } : {}"
        />

        <!-- 关闭 -->
        <button class="card-close" @click="emit('close')">✕</button>

        <!-- 主体 -->
        <div v-if="loading" class="card-loading">加载中…</div>
        <template v-else-if="user">
          <!-- 头像 + 基本信息 -->
          <div class="card-profile">
            <Avatar
              :src="getImgSrc(user.avatar)"
              :name="user.nickname"
              :size="64"
              :status="user.status"
              show-status
              class="card-avatar"
            />
            <div class="card-info">
              <div class="card-nickname">{{ user.nickname }}</div>
              <div class="card-username">{{ user.username }}</div>
              <div v-if="user.bio" class="card-bio">{{ user.bio }}</div>
            </div>
          </div>

          <!-- 待定信息区 -->
          <div class="card-meta">
            <div class="meta-item">
              <span class="meta-label">地区</span>
              <span class="meta-value">—</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">生日</span>
              <span class="meta-value">—</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">加入时间</span>
              <span class="meta-value">{{ new Date(user.created_at).toLocaleDateString() }}</span>
            </div>
          </div>

          <!-- 空间预览 -->
          <div class="space-preview">
            <div class="preview-header" @click="goToSpace">
              <span class="preview-title">✨ QQ 空间</span>
              <span class="preview-link">进入空间 →</span>
            </div>
            <div v-if="posts.length === 0" class="preview-empty">
              暂无动态
            </div>
            <div v-else class="preview-posts">
              <div
                v-for="post in posts"
                :key="post.id"
                class="preview-post"
                @click="goToSpace"
              >
                <!-- 文字摘要 -->
                <div class="preview-post-text">{{ post.content }}</div>
                <!-- 首图缩略 -->
                <img
                  v-if="parseImages(post.images)[0]"
                  :src="getImgSrc(parseImages(post.images)[0])"
                  class="preview-post-img"
                />
                <div class="preview-post-time">{{ timeAgo(post.created_at) }}</div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.card-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.user-card {
  width: 380px;
  background: var(--bg-surface);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
  overflow: hidden;
  position: relative;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

/* 封面 */
.card-cover {
  height: 100px;
  background: linear-gradient(135deg, #1677FF 0%, #52C41A 100%);
  background-size: cover;
  background-position: center;
  flex-shrink: 0;
}

.card-close {
  position: absolute;
  top: 8px;
  right: 10px;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(0,0,0,0.35);
  color: white;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
}
.card-close:hover { background: rgba(0,0,0,0.55); }

.card-loading {
  padding: 40px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
}

/* 头像悬出封面 */
.card-profile {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  padding: 0 20px 16px;
  margin-top: -32px;
  flex-shrink: 0;
}

.card-avatar {
  border: 3px solid white;
  border-radius: var(--radius-avatar);
  flex-shrink: 0;
}

.card-info {
  padding-bottom: 4px;
  min-width: 0;
}

.card-nickname {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-username {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.card-bio {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 待定信息 */
.card-meta {
  display: flex;
  gap: 0;
  padding: 12px 20px;
  border-top: 1px solid var(--border-light);
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.meta-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
}

.meta-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

.meta-value {
  font-size: 13px;
  color: var(--text-primary);
}

/* 空间预览 */
.space-preview {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px 8px;
  cursor: pointer;
  flex-shrink: 0;
}
.preview-header:hover .preview-link { text-decoration: underline; }

.preview-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.preview-link {
  font-size: 12px;
  color: var(--color-primary);
}

.preview-empty {
  padding: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
}

.preview-posts {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 0 0 8px;
}

.preview-post {
  padding: 10px 20px;
  cursor: pointer;
  transition: background 0.12s;
  display: grid;
  grid-template-columns: 1fr auto;
  grid-template-rows: auto auto;
  column-gap: 10px;
  row-gap: 4px;
}
.preview-post:hover { background: var(--bg-hover); }

.preview-post-text {
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.5;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  grid-column: 1;
  grid-row: 1;
}

.preview-post-img {
  width: 52px;
  height: 52px;
  border-radius: 6px;
  object-fit: cover;
  grid-column: 2;
  grid-row: 1 / 3;
  align-self: center;
}

.preview-post-time {
  font-size: 11px;
  color: var(--text-tertiary);
  grid-column: 1;
  grid-row: 2;
}

.preview-post-text::after {
  content: attr(data-time);
}
</style>
