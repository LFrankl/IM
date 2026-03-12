<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue'
import { onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import { useContactsStore } from '@/stores/contacts'
import { useGroupStore } from '@/stores/group'
import { useWS } from '@/composables/useWS'
import type { WSMessage, Message } from '@/types/chat'
import type { FriendRequest } from '@/types/user'

const auth = useAuthStore()
const chat = useChatStore()
const contacts = useContactsStore()
const group = useGroupStore()
const ws = useWS()

let offMessage: (() => void) | null = null

onMounted(async () => {
  if (!auth.token) return

  ws.connect(auth.token)

  offMessage = ws.onMessage((msg: WSMessage) => {
    switch (msg.type) {
      case 'message':
        if ((msg.data as Message).chat_type === 'group') {
          group.receiveMessage(msg.data as Message)
        } else {
          chat.receiveMessage(msg.data as Message)
        }
        break
      case 'message_sent':
        if ((msg.data as Message).chat_type === 'group') {
          group.confirmGroupSent(msg.data as Message)
        } else {
          chat.confirmSent(msg.data as Message)
        }
        break
      case 'friend_request':
        contacts.addPendingRequest((msg.data as { request: FriendRequest }).request)
        break
      case 'friend_accepted':
        contacts.fetchFriends()
        break
      case 'group_invite':
        group.addPendingInvite()
        break
    }
  })

  await Promise.all([
    chat.fetchConversations(),
    contacts.fetchPendingCount(),
    group.fetchMyGroups(),
    group.fetchPendingInviteCount(),
  ])
})

onUnmounted(() => {
  offMessage?.()
  ws.disconnect()
})
</script>

<template>
  <AppLayout />
</template>
