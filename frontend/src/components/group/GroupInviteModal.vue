<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { groupApi } from '@/api/group'
import type { GroupInvite } from '@/types/group'
import type { ApiResponse } from '@/api/client'

const emit = defineEmits<{ close: []; accepted: [] }>()

const invites = ref<GroupInvite[]>([])
const handling = ref<number | null>(null)

async function load() {
  const res = await groupApi.listMyInvites()
  const body = res.data as unknown as ApiResponse<GroupInvite[]>
  invites.value = body.data ?? []
}

async function handle(invite: GroupInvite, accept: boolean) {
  handling.value = invite.id
  try {
    await groupApi.handleInvite(invite.id, accept)
    invites.value = invites.value.filter(i => i.id !== invite.id)
    if (accept) emit('accepted')
  } finally {
    handling.value = null
  }
}

onMounted(load)
</script>

<template>
  <Teleport to="body">
    <div class="mask" @click.self="emit('close')">
      <div class="dialog">
        <div class="dialog-header">
          <span class="dialog-title">群邀请</span>
          <button class="close-btn" @click="emit('close')">✕</button>
        </div>
        <div class="invite-list">
          <div v-if="invites.length === 0" class="empty-tip">暂无邀请</div>
          <div v-for="inv in invites" :key="inv.id" class="invite-item">
            <div class="invite-info">
              <span class="group-name">{{ inv.group?.name }}</span>
              <span class="inviter-name">{{ inv.inviter?.nickname }} 邀请你加入</span>
            </div>
            <div class="invite-actions">
              <button
                class="accept-btn"
                :disabled="handling === inv.id"
                @click="handle(inv, true)"
              >接受</button>
              <button
                class="reject-btn"
                :disabled="handling === inv.id"
                @click="handle(inv, false)"
              >拒绝</button>
            </div>
          </div>
        </div>
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
  width: 360px;
  max-height: 70vh;
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

.invite-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.empty-tip {
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 40px 0;
}

.invite-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  gap: 12px;
}
.invite-item + .invite-item {
  border-top: 1px solid var(--border-light);
}

.invite-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.group-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.inviter-name {
  font-size: 12px;
  color: var(--text-secondary);
}

.invite-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.accept-btn {
  background: var(--qq-blue-primary);
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  padding: 4px 12px;
  cursor: pointer;
}
.accept-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.accept-btn:not(:disabled):hover { opacity: 0.88; }

.reject-btn {
  background: none;
  color: var(--text-secondary);
  border: 1px solid var(--border-normal);
  border-radius: 4px;
  font-size: 12px;
  padding: 4px 12px;
  cursor: pointer;
}
.reject-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.reject-btn:not(:disabled):hover { background: var(--bg-hover); }
</style>
