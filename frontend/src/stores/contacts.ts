import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Friendship, FriendRequest, FriendGroup } from '@/types/user'
import { friendApi } from '@/api/friend'
import type { ApiResponse } from '@/api/client'

export const useContactsStore = defineStore('contacts', () => {
  const friendships = ref<Friendship[]>([])
  const pendingRequests = ref<FriendRequest[]>([])
  const pendingCount = ref(0)
  const onlineIds = ref<Set<number>>(new Set())

  // 按分组聚合
  const friendGroups = computed<FriendGroup[]>(() => {
    const map = new Map<string, Friendship[]>()
    for (const f of friendships.value) {
      const g = f.group_name || '我的好友'
      if (!map.has(g)) map.set(g, [])
      map.get(g)!.push(f)
    }
    return Array.from(map.entries()).map(([name, friends]) => ({
      name,
      friends,
      collapsed: false,
    }))
  })

  async function fetchFriends() {
    const res = await friendApi.list()
    const body = res.data as unknown as ApiResponse<Friendship[]>
    friendships.value = body.data ?? []
  }

  async function fetchRequests() {
    const res = await friendApi.listRequests()
    const body = res.data as unknown as ApiResponse<FriendRequest[]>
    pendingRequests.value = body.data ?? []
    pendingCount.value = pendingRequests.value.length
  }

  async function fetchPendingCount() {
    const res = await friendApi.countPending()
    const body = res.data as unknown as ApiResponse<{ count: number }>
    pendingCount.value = body.data?.count ?? 0
  }

  function addPendingRequest(req: FriendRequest) {
    if (!pendingRequests.value.find((r) => r.id === req.id)) {
      pendingRequests.value.unshift(req)
      pendingCount.value++
    }
  }

  function removePendingRequest(reqId: number) {
    pendingRequests.value = pendingRequests.value.filter((r) => r.id !== reqId)
    pendingCount.value = Math.max(0, pendingCount.value - 1)
  }

  function addFriend(friendship: Friendship) {
    if (!friendships.value.find((f) => f.friend_id === friendship.friend_id)) {
      friendships.value.push(friendship)
    }
  }

  function removeFriend(friendId: number) {
    friendships.value = friendships.value.filter((f) => f.friend_id !== friendId)
  }

  function setOnline(userId: number) {
    onlineIds.value = new Set([...onlineIds.value, userId])
  }

  function setOffline(userId: number) {
    const next = new Set(onlineIds.value)
    next.delete(userId)
    onlineIds.value = next
  }

  function isOnline(userId: number): boolean {
    return onlineIds.value.has(userId)
  }

  return {
    friendships,
    pendingRequests,
    pendingCount,
    onlineIds,
    friendGroups,
    fetchFriends,
    fetchRequests,
    fetchPendingCount,
    addPendingRequest,
    removePendingRequest,
    addFriend,
    removeFriend,
    setOnline,
    setOffline,
    isOnline,
  }
})
