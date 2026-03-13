<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Message } from '@/types/chat'
import { useAuthStore } from '@/stores/auth'
import { useUserCard } from '@/composables/useUserCard'
import Avatar from '@/components/common/Avatar.vue'

const props = defineProps<{ msg: Message; showName?: boolean }>()
defineEmits<{ openCard: [userId: number] }>()
const auth = useAuthStore()
const { openCard } = useUserCard()

const isSelf = computed(() => Number(props.msg.from_id) === Number(auth.user?.id))

// content 字段在数据库/WS 里始终是 JSON 字符串，统一解析
const parsedContent = computed<Record<string, any>>(() => {
  try {
    return typeof props.msg.content === 'string'
      ? JSON.parse(props.msg.content)
      : (props.msg.content as any) ?? {}
  } catch {
    return {}
  }
})

const textContent = computed(() => {
  if (props.msg.msg_type !== 'text') return ''
  return parsedContent.value.text ?? String(props.msg.content)
})

function getMediaSrc(url: string | undefined) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

const timeStr = computed(() => {
  const d = new Date(props.msg.created_at)
  const h = d.getHours().toString().padStart(2, '0')
  const m = d.getMinutes().toString().padStart(2, '0')
  return `${h}:${m}`
})

function getAvatarSrc(url: string | undefined) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

const previewSrc = ref<string | null>(null)

function openPreview(src: string) {
  previewSrc.value = src
}

function closePreview() {
  previewSrc.value = null
}

function onPreviewKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closePreview()
}
</script>

<template>
  <div class="msg-row" :class="{ self: isSelf }">
    <!-- 对方头像 -->
    <Avatar
      v-if="!isSelf"
      :src="getAvatarSrc(msg.from?.avatar)"
      :name="msg.from?.nickname"
      :size="36"
      style="cursor:pointer"
      @click="openCard(Number(msg.from_id))"
    />

    <div class="msg-body">
      <!-- 发送者名（群聊时显示） -->
      <div v-if="showName && !isSelf" class="msg-sender">
        {{ msg.from?.nickname }}
      </div>

      <div class="bubble-wrap">
        <div class="bubble" :class="{ 'bubble-self': isSelf, 'bubble-other': !isSelf }">
          <template v-if="msg.msg_type === 'text'">
            <span class="selectable">{{ textContent }}</span>
          </template>
          <template v-else-if="msg.msg_type === 'image'">
            <img
              :src="getMediaSrc(parsedContent.url)"
              class="msg-img"
              @click="openPreview(getMediaSrc(parsedContent.url))"
            />
          </template>
          <template v-else-if="msg.msg_type === 'file'">
            <div class="msg-file">
              <span class="file-icon">📎</span>
              <a :href="getMediaSrc(parsedContent.url)" target="_blank" class="file-name">
                {{ parsedContent.name }}
              </a>
            </div>
          </template>
        </div>
        <span class="msg-time">{{ timeStr }}</span>
      </div>
    </div>

    <!-- 自己头像 -->
    <Avatar
      v-if="isSelf"
      :src="getAvatarSrc(auth.user?.avatar)"
      :name="auth.user?.nickname"
      :size="36"
    />
  </div>

  <!-- 图片预览遮罩 -->
  <Teleport to="body">
    <div
      v-if="previewSrc"
      class="img-preview-mask"
      @click="closePreview"
      @keydown="onPreviewKeydown"
      tabindex="0"
    >
      <img :src="previewSrc" class="img-preview-full" @click.stop />
      <button class="img-preview-close" @click="closePreview">✕</button>
    </div>
  </Teleport>
</template>

<style scoped>
.msg-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 16px;
}

.msg-row.self {
  justify-content: flex-end;
}

.msg-body {
  max-width: 60%;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.msg-row.self .msg-body {
  align-items: flex-end;
}

.msg-sender {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 0 4px;
}

.bubble-wrap {
  display: flex;
  align-items: flex-end;
  gap: 6px;
}

.msg-row.self .bubble-wrap {
  flex-direction: row-reverse;
}

.bubble {
  padding: 10px 14px;
  border-radius: var(--radius-bubble);
  font-size: 14px;
  line-height: 1.6;
  word-break: break-all;
}

.bubble-other {
  background: var(--color-bubble-other);
  box-shadow: var(--shadow-bubble-other);
  border-radius: 2px 8px 8px 8px;
}

.bubble-self {
  background: var(--color-bubble-self);
  border-radius: 8px 2px 8px 8px;
}

.msg-time {
  font-size: 11px;
  color: var(--text-tertiary);
  flex-shrink: 0;
  padding-bottom: 2px;
}

.msg-img {
  max-width: 200px;
  max-height: 200px;
  border-radius: 4px;
  cursor: zoom-in;
  display: block;
}

.msg-file {
  display: flex;
  align-items: center;
  gap: 6px;
}

.file-name {
  font-size: 13px;
  color: var(--text-link);
  text-decoration: underline;
  cursor: pointer;
}

/* ── 图片预览 ── */
.img-preview-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  cursor: zoom-out;
  animation: mask-in 0.18s ease;
}

@keyframes mask-in {
  from { opacity: 0; }
  to   { opacity: 1; }
}

.img-preview-full {
  max-width: 90vw;
  max-height: 90vh;
  border-radius: 6px;
  object-fit: contain;
  cursor: default;
  box-shadow: 0 8px 48px rgba(0, 0, 0, 0.6);
  animation: img-in 0.2s ease;
}

@keyframes img-in {
  from { transform: scale(0.92); opacity: 0; }
  to   { transform: scale(1);    opacity: 1; }
}

.img-preview-close {
  position: fixed;
  top: 20px;
  right: 24px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}
.img-preview-close:hover { background: rgba(255, 255, 255, 0.28); }
</style>
