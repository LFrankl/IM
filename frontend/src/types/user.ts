export type UserStatus = 'online' | 'offline' | 'busy'

export interface User {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  status: UserStatus
  created_at: string
}

export interface Friendship {
  id: number
  user_id: number
  friend_id: number
  remark: string
  group_name: string
  created_at: string
  friend: User
}

export interface FriendRequest {
  id: number
  from_id: number
  to_id: number
  message: string
  status: 'pending' | 'accepted' | 'rejected'
  created_at: string
  from: User
}

export interface FriendGroup {
  name: string
  friends: Friendship[]
  collapsed: boolean
}
