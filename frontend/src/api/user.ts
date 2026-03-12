import client from './client'
import type { User } from '@/types/user'

export const userApi = {
  search(q: string) {
    return client.get<User[]>('/api/users/search', { params: { q } })
  },
  getById(id: number) {
    return client.get<User>(`/api/users/${id}`)
  },
  uploadAvatar(file: File) {
    const form = new FormData()
    form.append('avatar', file)
    return client.post<User>('/api/users/me/avatar', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  updateProfile(nickname: string, bio: string) {
    return client.put<User>('/api/users/me', { nickname, bio })
  },
}
