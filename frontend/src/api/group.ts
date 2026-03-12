import client from './client'
import type { Group, GroupMember } from '@/types/group'

export const groupApi = {
  list() {
    return client.get<Group[]>('/api/groups')
  },
  create(name: string, memberIds: number[]) {
    return client.post<Group>('/api/groups', { name, member_ids: memberIds })
  },
  search(q: string) {
    return client.get<Group[]>('/api/groups/search', { params: { q } })
  },
  get(id: number) {
    return client.get<{ group: Group; members: GroupMember[] }>(`/api/groups/${id}`)
  },
  getMembers(id: number) {
    return client.get<GroupMember[]>(`/api/groups/${id}/members`)
  },
  join(id: number) {
    return client.post(`/api/groups/${id}/join`)
  },
  leave(id: number) {
    return client.delete(`/api/groups/${id}/leave`)
  },
  kickMember(groupId: number, userId: number) {
    return client.delete(`/api/groups/${groupId}/members/${userId}`)
  },
  disband(id: number) {
    return client.delete(`/api/groups/${id}`)
  },
  getMessages(id: number, beforeId?: number, limit = 30) {
    return client.get(`/api/groups/${id}/messages`, {
      params: { before_id: beforeId || undefined, limit },
    })
  },
}
