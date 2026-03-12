<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { spaceApi, type SpacePost, type SpaceComment } from '@/api/space'
import { userApi } from '@/api/user'
import type { ApiResponse } from '@/api/client'
import type { User } from '@/types/user'
import Avatar from '@/components/common/Avatar.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

// 当前访问的空间主人 ID，没有则看自己
const targetUserId = computed(() => {
  const p = route.params.userId
  return p ? Number(p) : auth.user?.id ?? 0
})

const isSelf = computed(() => targetUserId.value === auth.user?.id)

const posts = ref<SpacePost[]>([])
const loading = ref(false)
const noMore = ref(false)
const newContent = ref('')
const posting = ref(false)
const commentInputs = ref<Record<number, string>>({})
const expandedComments = ref<Set<number>>(new Set())

// 切换空间主人时重置
watch(targetUserId, () => {
  posts.value = []
  noMore.value = false
  loadPosts()
}, { immediate: true })

async function loadPosts(before?: number) {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = isSelf.value
      ? await spaceApi.getFeed(before)
      : await spaceApi.getUserPosts(targetUserId.value, before)
    const body = res.data as unknown as ApiResponse<SpacePost[]>
    const newPosts = body.data ?? []
    if (before) {
      posts.value = [...posts.value, ...newPosts]
    } else {
      posts.value = newPosts
    }
    if (newPosts.length < 20) noMore.value = true
  } finally {
    loading.value = false
  }
}

async function submitPost() {
  if (!newContent.value.trim()) return
  posting.value = true
  try {
    const res = await spaceApi.createPost(newContent.value.trim())
    const body = res.data as unknown as ApiResponse<SpacePost>
    if (body.data) {
      posts.value = [{ ...body.data, comments: [] }, ...posts.value]
      newContent.value = ''
    }
  } finally {
    posting.value = false
  }
}

async function deletePost(post: SpacePost) {
  if (!confirm('确定删除这条动态？')) return
  await spaceApi.deletePost(post.id)
  posts.value = posts.value.filter((p) => p.id !== post.id)
}

async function toggleLike(post: SpacePost) {
  if (post.liked) {
    await spaceApi.unlikePost(post.id)
    post.liked = false
    post.like_count--
  } else {
    await spaceApi.likePost(post.id)
    post.liked = true
    post.like_count++
  }
}

function toggleComments(postId: number) {
  if (expandedComments.value.has(postId)) {
    expandedComments.value.delete(postId)
  } else {
    expandedComments.value.add(postId)
  }
  expandedComments.value = new Set(expandedComments.value)
}

async function submitComment(post: SpacePost) {
  const content = (commentInputs.value[post.id] ?? '').trim()
  if (!content) return
  const res = await spaceApi.addComment(post.id, content)
  const body = res.data as unknown as ApiResponse<SpaceComment>
  if (body.data) {
    if (!post.comments) post.comments = []
    post.comments.push(body.data)
    commentInputs.value[post.id] = ''
    expandedComments.value.add(post.id)
    expandedComments.value = new Set(expandedComments.value)
  }
}

async function deleteComment(post: SpacePost, comment: SpaceComment) {
  await spaceApi.deleteComment(comment.id)
  post.comments = post.comments?.filter((c) => c.id !== comment.id)
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
  if (d < 30) return `${d}天前`
  return new Date(dateStr).toLocaleDateString()
}

const coverUploading = ref(false)

async function onCoverChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  coverUploading.value = true
  try {
    const res = await userApi.uploadCover(file)
    const body = res.data as unknown as ApiResponse<User>
    auth.updateUser(body.data)
  } finally {
    coverUploading.value = false
    ;(e.target as HTMLInputElement).value = ''
  }
}

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

function goToSpace(userId: number) {
  router.push({ name: 'space', params: { userId } })
}
</script>

<template>
  <div class="space-view">
    <!-- 左侧内容 -->
    <div class="space-main">
      <!-- 顶部封面 + 个人信息 -->
      <div class="space-header">
        <div
          class="cover-bg"
          :style="auth.user?.cover && isSelf ? { backgroundImage: `url(${getAvatarSrc(auth.user.cover)})` } : {}"
          @click="isSelf && ($refs.coverInput as HTMLInputElement).click()"
        >
          <div v-if="isSelf" class="cover-upload-hint">
            <span>{{ coverUploading ? '上传中…' : '点击更换封面' }}</span>
          </div>
          <input
            ref="coverInput"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            style="display:none"
            @change="onCoverChange"
          />
        </div>
        <div class="profile-info">
          <Avatar
            class="profile-avatar"
            :src="getAvatarSrc(isSelf ? auth.user?.avatar : posts[0]?.user?.avatar)"
            :name="isSelf ? auth.user?.nickname : (posts[0]?.user?.nickname ?? '用户')"
            :size="80"
          />
          <div class="profile-text">
            <div class="profile-name">
              {{ isSelf ? auth.user?.nickname : posts[0]?.user?.nickname ?? '用户空间' }}
            </div>
            <div class="profile-bio">{{ auth.user?.bio || (isSelf ? '这个人很懒，什么都没写～' : '') }}</div>
          </div>
        </div>
      </div>

      <!-- 发帖框（仅自己的空间） -->
      <div v-if="isSelf" class="post-composer">
        <textarea
          v-model="newContent"
          placeholder="分享你的动态…"
          rows="3"
          class="composer-input"
          @keydown.ctrl.enter="submitPost"
        />
        <div class="composer-footer">
          <span class="composer-tip">Ctrl+Enter 发布</span>
          <button class="post-btn" :disabled="!newContent.trim() || posting" @click="submitPost">
            {{ posting ? '发布中…' : '发布' }}
          </button>
        </div>
      </div>

      <!-- 动态列表 -->
      <div class="posts-list">
        <div v-for="post in posts" :key="post.id" class="post-card">
          <!-- 帖子头部 -->
          <div class="post-header">
            <Avatar
              class="post-avatar"
              :src="getAvatarSrc(post.user?.avatar)"
              :name="post.user?.nickname"
              :size="40"
              @click="goToSpace(post.user_id)"
            />
            <div class="post-meta">
              <span class="post-author" @click="goToSpace(post.user_id)">
                {{ post.user?.nickname }}
              </span>
              <span class="post-time">{{ timeAgo(post.created_at) }}</span>
            </div>
            <button
              v-if="post.user_id === auth.user?.id"
              class="delete-post-btn"
              @click="deletePost(post)"
              title="删除"
            >✕</button>
          </div>

          <!-- 正文 -->
          <div class="post-content">{{ post.content }}</div>

          <!-- 图片 -->
          <div v-if="parseImages(post.images).length > 0" class="post-images">
            <img
              v-for="(img, i) in parseImages(post.images)"
              :key="i"
              :src="img"
              class="post-image"
            />
          </div>

          <!-- 操作栏 -->
          <div class="post-actions">
            <button
              class="action-btn"
              :class="{ liked: post.liked }"
              @click="toggleLike(post)"
            >
              {{ post.liked ? '❤️' : '🤍' }} {{ post.like_count }}
            </button>
            <button class="action-btn" @click="toggleComments(post.id)">
              💬 {{ post.comments?.length ?? 0 }}
            </button>
          </div>

          <!-- 评论区 -->
          <div v-if="expandedComments.has(post.id)" class="comments-section">
            <div
              v-for="comment in post.comments"
              :key="comment.id"
              class="comment-item"
            >
              <span
                class="comment-author"
                @click="goToSpace(comment.user_id)"
              >{{ comment.user?.nickname }}</span>
              <span class="comment-content">：{{ comment.content }}</span>
              <button
                v-if="comment.user_id === auth.user?.id || post.user_id === auth.user?.id"
                class="delete-comment-btn"
                @click="deleteComment(post, comment)"
              >✕</button>
            </div>
            <div class="comment-input-row">
              <input
                v-model="commentInputs[post.id]"
                placeholder="写评论…"
                class="comment-input"
                @keydown.enter="submitComment(post)"
              />
              <button class="comment-submit" @click="submitComment(post)">发送</button>
            </div>
          </div>
        </div>

        <!-- 加载更多 -->
        <div v-if="posts.length > 0 && !noMore" class="load-more">
          <button @click="loadPosts(posts[posts.length - 1]?.id)" :disabled="loading">
            {{ loading ? '加载中…' : '加载更多' }}
          </button>
        </div>
        <div v-if="noMore && posts.length > 0" class="no-more">没有更多了</div>
        <div v-if="!loading && posts.length === 0" class="empty-posts">
          {{ isSelf ? '还没有动态，发布第一条吧～' : '该用户还没有动态' }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.space-view {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  background: #f0f2f5;
}

.space-main {
  max-width: 680px;
  margin: 0 auto;
  padding-bottom: 40px;
}

/* 头部 */
.space-header {
  position: relative;
  margin-bottom: 16px;
}

.cover-bg {
  height: 180px;
  background: linear-gradient(135deg, #1677FF 0%, #52C41A 100%);
  background-size: cover;
  background-position: center;
  border-radius: 0 0 12px 12px;
  position: relative;
  cursor: pointer;
  overflow: hidden;
}

.cover-upload-hint {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.3);
  opacity: 0;
  transition: opacity 0.2s;
}

.cover-bg:hover .cover-upload-hint {
  opacity: 1;
}

.cover-upload-hint span {
  color: white;
  font-size: 14px;
  background: rgba(0,0,0,0.4);
  padding: 6px 16px;
  border-radius: 20px;
}

.profile-info {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  padding: 0 24px;
  margin-top: -36px;
  position: relative;
  z-index: 1;
}

.profile-avatar {
  border: 4px solid white;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
  flex-shrink: 0;
}

.profile-text {
  padding-bottom: 8px;
}

.profile-name {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.profile-bio {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
}

/* 发帖框 */
.post-composer {
  background: white;
  border-radius: 12px;
  padding: 16px;
  margin: 0 0 16px;
  box-shadow: var(--shadow-card);
}

.composer-input {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-primary);
  font-family: inherit;
  box-sizing: border-box;
}

.composer-input::placeholder { color: var(--text-tertiary); }

.composer-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border-light);
}

.composer-tip { font-size: 12px; color: var(--text-tertiary); }

.post-btn {
  background: var(--qq-blue-primary);
  color: white;
  border: none;
  border-radius: 6px;
  padding: 6px 20px;
  font-size: 13px;
  cursor: pointer;
}
.post-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.post-btn:not(:disabled):hover { opacity: 0.88; }

/* 帖子列表 */
.posts-list { display: flex; flex-direction: column; gap: 12px; }

.post-card {
  background: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: var(--shadow-card);
}

.post-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.post-avatar {
  cursor: pointer;
  flex-shrink: 0;
}

.post-meta { flex: 1; }

.post-author {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  cursor: pointer;
}
.post-author:hover { color: var(--qq-blue-primary); }

.post-time { font-size: 12px; color: var(--text-tertiary); }

.delete-post-btn {
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 13px;
  padding: 4px;
  border-radius: 4px;
}
.delete-post-btn:hover { background: var(--bg-hover); color: var(--color-error); }

.post-content {
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 10px;
}

.post-images {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.post-image {
  max-width: 200px;
  max-height: 200px;
  border-radius: 8px;
  object-fit: cover;
  cursor: pointer;
}

/* 操作栏 */
.post-actions {
  display: flex;
  gap: 16px;
  padding-top: 10px;
  border-top: 1px solid var(--border-light);
}

.action-btn {
  background: none;
  border: none;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.12s;
}
.action-btn:hover { background: var(--bg-hover); }
.action-btn.liked { color: #eb2f96; }

/* 评论区 */
.comments-section {
  margin-top: 10px;
  padding: 10px 12px;
  background: #f7f8fa;
  border-radius: 8px;
}

.comment-item {
  font-size: 13px;
  line-height: 1.6;
  padding: 3px 0;
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.comment-author {
  font-weight: 600;
  color: var(--qq-blue-primary);
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
}
.comment-author:hover { text-decoration: underline; }

.comment-content {
  color: var(--text-primary);
  flex: 1;
  word-break: break-word;
}

.delete-comment-btn {
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 11px;
  padding: 0 3px;
  flex-shrink: 0;
}
.delete-comment-btn:hover { color: var(--color-error); }

.comment-input-row {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.comment-input {
  flex: 1;
  height: 32px;
  padding: 0 10px;
  border: 1px solid var(--border-input);
  border-radius: 16px;
  font-size: 13px;
  outline: none;
  user-select: text;
}
.comment-input:focus { border-color: var(--qq-blue-primary); }

.comment-submit {
  height: 32px;
  padding: 0 14px;
  background: var(--qq-blue-primary);
  color: white;
  border: none;
  border-radius: 16px;
  font-size: 13px;
  cursor: pointer;
}
.comment-submit:hover { opacity: 0.88; }

/* 底部 */
.load-more {
  text-align: center;
  padding: 12px 0;
}
.load-more button {
  background: none;
  border: 1px solid var(--border-normal);
  border-radius: 20px;
  padding: 6px 24px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}
.load-more button:hover { background: var(--bg-hover); }
.load-more button:disabled { opacity: 0.5; cursor: not-allowed; }

.no-more {
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 12px 0;
}

.empty-posts {
  text-align: center;
  font-size: 14px;
  color: var(--text-tertiary);
  padding: 48px 0;
}
</style>
