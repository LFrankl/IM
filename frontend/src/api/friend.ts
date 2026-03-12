import client from './client'
import type { Friendship, FriendRequest } from '@/types/user'

export const friendApi = {
  list() {
    return client.get<Friendship[]>('/api/friends')
  },
  sendRequest(toId: number, message = '') {
    return client.post<FriendRequest>('/api/friends/requests', { to_id: toId, message })
  },
  listRequests() {
    return client.get<FriendRequest[]>('/api/friends/requests')
  },
  countPending() {
    return client.get<{ count: number }>('/api/friends/requests/count')
  },
  handleRequest(id: number, action: 'accept' | 'reject') {
    return client.put(`/api/friends/requests/${id}`, { action })
  },
  deleteFriend(id: number) {
    return client.delete(`/api/friends/${id}`)
  },
  updateRemark(id: number, remark: string) {
    return client.put(`/api/friends/${id}/remark`, { remark })
  },
  updateGroup(id: number, groupName: string) {
    return client.put(`/api/friends/${id}/group`, { group_name: groupName })
  },
}
