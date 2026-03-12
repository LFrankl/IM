<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
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
const region = ref(auth.user?.region ?? '')
const birthday = ref(auth.user?.birthday ?? '')
const saving = ref(false)
const errorMsg = ref('')

// 头像：暂存本地预览，保存时才上传
const pendingAvatarFile = ref<File | null>(null)
const previewUrl = ref<string | null>(null)

function getAvatarSrc(url: string) {
  if (!url) return undefined
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}

const displayAvatar = ref(getAvatarSrc(auth.user?.avatar ?? '') ?? '')

function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  ;(e.target as HTMLInputElement).value = ''
  // 释放旧的预览 URL
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = URL.createObjectURL(file)
  pendingAvatarFile.value = file
  displayAvatar.value = previewUrl.value
}

async function save() {
  if (!nickname.value.trim()) {
    errorMsg.value = '昵称不能为空'
    return
  }
  saving.value = true
  errorMsg.value = ''
  try {
    // 如果有新头像，先上传
    if (pendingAvatarFile.value) {
      const res = await userApi.uploadAvatar(pendingAvatarFile.value)
      const body = res.data as unknown as ApiResponse<User>
      auth.updateUser(body.data)
    }
    // 更新昵称/签名/地区/生日
    const res = await userApi.updateProfile(nickname.value.trim(), bio.value.trim(), region.value.trim(), birthday.value)
    const body = res.data as unknown as ApiResponse<User>
    auth.updateUser(body.data)
    emit('close')
  } catch {
    errorMsg.value = '保存失败，请重试'
  } finally {
    saving.value = false
  }
}

onUnmounted(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
})
</script>

<template>
  <Modal title="编辑资料" :width="400" max-height="92vh" @close="emit('close')">
    <div class="profile-modal">
      <!-- 头像区 -->
      <div class="avatar-section">
        <div class="avatar-wrap">
          <Avatar
            :src="displayAvatar || undefined"
            :name="auth.user?.nickname"
            :size="72"
          />
          <div class="avatar-overlay" @click="($refs.fileInput as HTMLInputElement).click()">
            <span>更换</span>
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
      <div class="form-row">
        <div class="form-group">
          <label>地区</label>
          <input v-model="region" class="form-input" placeholder="城市" maxlength="30" />
        </div>
        <div class="form-group">
          <label>生日</label>
          <input v-model="birthday" class="form-input" type="text" placeholder="YYYY-MM-DD" maxlength="10" />
        </div>
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
  width: 100%;
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

.form-row {
  display: flex;
  gap: 12px;
}

.form-row .form-group {
  flex: 1;
  min-width: 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.form-input {
  height: 34px;
  padding: 0 0 6px;
  border: none;
  border-bottom: 1px solid var(--border-light);
  border-radius: 0;
  font-size: 14px;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  transition: border-color 0.2s;
  width: 100%;
  box-sizing: border-box;
}

.form-input:focus {
  border-bottom-color: var(--qq-blue-primary);
}

.error-msg {
  font-size: 13px;
  color: var(--color-error);
  margin: 0;
}

.modal-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 4px;
}

.btn-cancel {
  padding: 6px 16px;
  border: 1px solid var(--border-normal);
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary);
}
.btn-cancel:hover { background: var(--bg-hover); }

.btn-save {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  background: var(--qq-blue-primary);
  color: white;
  cursor: pointer;
  font-size: 14px;
}

.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-save:not(:disabled):hover { opacity: 0.88; }
</style>
