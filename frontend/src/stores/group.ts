import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Group, GroupMember, GroupInvite } from '@/types/group'
import type { Message } from '@/types/chat'
import { groupApi } from '@/api/group'
import type { ApiResponse } from '@/api/client'

export const useGroupStore = defineStore('group', () => {
  const myGroups = ref<GroupWithMeta[]>([])
  const activeGroupId = ref<number | null>(null)
  const membersCache = ref<Record<number, GroupMember[]>>({})
  const messagesCache = ref<Record<number, Message[]>>({})
  const pendingInviteCount = ref(0)

  const totalUnread = computed(() =>
    myGroups.value.reduce((sum, g) => sum + (g.unread_count ?? 0), 0)
  )

  const activeGroup = () => myGroups.value.find((g) => g.id === activeGroupId.value) ?? null

  async function fetchMyGroups() {
    const res = await groupApi.list()
    const body = res.data as unknown as ApiResponse<GroupWithMeta[]>
    const existing = new Map(myGroups.value.map((g) => [g.id, g.unread_count ?? 0]))
    myGroups.value = (body.data ?? []).map((g) => ({ ...g, unread_count: existing.get(g.id) ?? 0 }))
  }

  async function fetchPendingInviteCount() {
    const res = await groupApi.listMyInvites()
    const body = res.data as unknown as ApiResponse<GroupInvite[]>
    pendingInviteCount.value = (body.data ?? []).length
  }

  function addPendingInvite() {
    pendingInviteCount.value++
  }

  function clearPendingInvites() {
    pendingInviteCount.value = 0
  }

  async function fetchMessages(groupId: number, beforeId?: number) {
    const res = await groupApi.getMessages(groupId, beforeId)
    const body = res.data as unknown as ApiResponse<Message[]>
    const msgs = body.data ?? []
    if (beforeId) {
      messagesCache.value[groupId] = [...msgs, ...(messagesCache.value[groupId] ?? [])]
    } else {
      messagesCache.value[groupId] = msgs
    }
    return msgs
  }

  async function fetchMembers(groupId: number) {
    const res = await groupApi.get(groupId)
    const body = res.data as unknown as ApiResponse<{ group: Group; members: GroupMember[] }>
    if (body.data) {
      membersCache.value[groupId] = body.data.members
    }
    return body.data?.members ?? []
  }

  function setActiveGroup(id: number | null) {
    activeGroupId.value = id
    // 切换群组时清零未读
    if (id !== null) {
      myGroups.value = myGroups.value.map((g) =>
        g.id === id ? { ...g, unread_count: 0 } : g
      )
    }
  }

  function receiveMessage(msg: Message) {
    const gid = Number(msg.to_id)
    if (!messagesCache.value[gid]) messagesCache.value[gid] = []
    if (!messagesCache.value[gid].find((m) => m.id === msg.id)) {
      messagesCache.value[gid] = [...messagesCache.value[gid], msg]
    }
    myGroups.value = myGroups.value.map((item) => {
      if (item.id !== gid) return item
      const unread = gid !== activeGroupId.value
        ? (item.unread_count ?? 0) + 1
        : (item.unread_count ?? 0)
      return { ...item, last_message: msg, unread_count: unread }
    })
  }

  function confirmGroupSent(msg: Message) {
    const gid = Number(msg.to_id)
    if (!messagesCache.value[gid]) messagesCache.value[gid] = []
    if (!messagesCache.value[gid].find((m) => m.id === msg.id)) {
      messagesCache.value[gid] = [...messagesCache.value[gid], msg]
    }
    myGroups.value = myGroups.value.map((item) =>
      item.id === gid ? { ...item, last_message: msg } : item
    )
  }

  function removeGroup(id: number) {
    myGroups.value = myGroups.value.filter((g) => g.id !== id)
    if (activeGroupId.value === id) activeGroupId.value = null
  }

  function updateGroupAvatar(groupId: number, avatarUrl: string) {
    myGroups.value = myGroups.value.map((g) =>
      g.id === groupId ? { ...g, avatar: avatarUrl } : g
    )
  }

  return {
    myGroups,
    activeGroupId,
    messagesCache,
    membersCache,
    pendingInviteCount,
    totalUnread,
    activeGroup,
    fetchMyGroups,
    fetchPendingInviteCount,
    addPendingInvite,
    clearPendingInvites,
    fetchMessages,
    fetchMembers,
    setActiveGroup,
    receiveMessage,
    confirmGroupSent,
    removeGroup,
    updateGroupAvatar,
  }
})

export interface GroupWithMeta extends Group {
  member_count?: number
  last_message?: Message | null
  unread_count?: number
}
