import type { User } from './user'

export interface Group {
  id: number
  name: string
  avatar: string
  notice: string
  owner_id: number
  created_at: string
  owner?: User
  members?: GroupMember[]
  member_count?: number
}

export interface GroupMember {
  id: number
  group_id: number
  user_id: number
  joined_at: string
  user?: User
}
