import client from './client'
import type { User } from '@/types/user'

export const userApi = {
  search(q: string) {
    return client.get<User[]>('/api/users/search', { params: { q } })
  },
  getById(id: number) {
    return client.get<User>(`/api/users/${id}`)
  },
}
