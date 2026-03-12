<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/user'
import type { ApiResponse } from '@/api/client'
import type { User } from '@/types/user'
import Avatar from '@/components/common/Avatar.vue'
import Modal from '@/components/common/Modal.vue'

const emit = defineEmits<{ (e: 'close'): void }>()

const auth = useAuthStore()

const nickname = ref(auth.user?.nickname ?? '')
const bio = ref(auth.user?.bio ?? '')
const uploading = ref(false)
const saving = ref(false)
const errorMsg = ref('')

const avatarUrl = ref(auth.user?.avatar ?? '')

async function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploading.value = true
  errorMsg.value = ''
  try {
    const res = await userApi.uploadAvatar(file)
    const body = res.data as unknown as ApiResponse<User>
    auth.updateUser(body.data)
    avatarUrl.value = body.data.avatar
  } catch {
    errorMsg.value = '上传失败，请重试'
  } finally {
    uploading.value = false
  }
}

async function save() {
  if (!nickname.value.trim()) {
    errorMsg.value = '昵称不能为空'
    return
  }
  saving.value = true
  errorMsg.value = ''
  try {
    const res = await userApi.updateProfile(nickname.value.trim(), bio.value.trim())
    const body = res.data as unknown as ApiResponse<User>
    auth.updateUser(body.data)
    emit('close')
  } catch {
    errorMsg.value = '保存失败，请重试'
  } finally {
    saving.value = false
  }
}

function getAvatarSrc(url: string) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}
</script>

<template>
  <Modal title="编辑资料" @close="emit('close')">
    <div class="profile-modal">
      <!-- 头像区 -->
      <div class="avatar-section">
        <div class="avatar-wrap">
          <Avatar
            :src="getAvatarSrc(avatarUrl)"
            :name="auth.user?.nickname"
            :size="72"
          />
          <div class="avatar-overlay" @click="($refs.fileInput as HTMLInputElement).click()">
            <span>{{ uploading ? '上传中…' : '更换' }}</span>
          </div>
        </div>
        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          style="display:none"
          @change="onFileChange"
        />
      </div>

      <!-- 表单 -->
      <div class="form-group">
        <label>昵称</label>
        <input v-model="nickname" class="form-input" placeholder="请输入昵称" maxlength="50" />
      </div>
      <div class="form-group">
        <label>个性签名</label>
        <input v-model="bio" class="form-input" placeholder="介绍一下自己…" maxlength="100" />
      </div>

      <p v-if="errorMsg" class="error-msg">{{ errorMsg }}</p>

      <div class="modal-actions">
        <button class="btn-cancel" @click="emit('close')">取消</button>
        <button class="btn-save" :disabled="saving" @click="save">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.profile-modal {
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.avatar-section {
  display: flex;
  justify-content: center;
}

.avatar-wrap {
  position: relative;
  cursor: pointer;
  border-radius: var(--radius-avatar);
  overflow: hidden;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
  border-radius: var(--radius-avatar);
}

.avatar-wrap:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay span {
  color: white;
  font-size: 13px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  color: var(--text-secondary);
}

.form-input {
  height: 36px;
  padding: 0 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: var(--color-primary);
}

.error-msg {
  font-size: 13px;
  color: var(--color-error);
  margin: 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

.btn-cancel {
  padding: 6px 16px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 14px;
}

.btn-save {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  background: var(--color-primary);
  color: white;
  cursor: pointer;
  font-size: 14px;
}

.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
