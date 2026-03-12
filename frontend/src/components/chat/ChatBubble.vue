<script setup lang="ts">
import { computed } from 'vue'
import type { Message } from '@/types/chat'
import { useAuthStore } from '@/stores/auth'
import Avatar from '@/components/common/Avatar.vue'

const props = defineProps<{ msg: Message; showName?: boolean }>()
const emit = defineEmits<{ openCard: [userId: number] }>()
const auth = useAuthStore()

const isSelf = computed(() => Number(props.msg.from_id) === Number(auth.user?.id))

const textContent = computed(() => {
  if (props.msg.msg_type !== 'text') return ''
  try {
    const c = typeof props.msg.content === 'string'
      ? JSON.parse(props.msg.content)
      : props.msg.content
    return (c as { text: string }).text
  } catch {
    return String(props.msg.content)
  }
})

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
      @click="emit('openCard', Number(msg.from_id))"
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
            <img :src="(msg.content as any).url" class="msg-img" />
          </template>
          <template v-else-if="msg.msg_type === 'file'">
            <div class="msg-file">
              <span class="file-icon">📎</span>
              <span class="file-name">{{ (msg.content as any).name }}</span>
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
  cursor: pointer;
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
</style>
