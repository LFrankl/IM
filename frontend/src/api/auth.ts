import client from './client'
import type { User } from '@/types/user'

export const authApi = {
  register(data: { username: string; password: string; nickname: string }) {
    return client.post<{ token: string; user: User }>('/api/auth/register', data)
  },
  login(data: { username: string; password: string }) {
    return client.post<{ token: string; user: User }>('/api/auth/login', data)
  },
  logout() {
    return client.post('/api/auth/logout')
  },
  me() {
    return client.get<User>('/api/auth/me')
  },
}
