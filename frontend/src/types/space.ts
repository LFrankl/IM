import type { User } from './user'

export interface SpacePost {
  id: number
  user_id: number
  content: string
  images: string[]
  like_count: number
  liked: boolean
  created_at: string
  updated_at: string
  user?: User
  comments?: SpaceComment[]
}

export interface SpaceComment {
  id: number
  post_id: number
  user_id: number
  content: string
  created_at: string
  user?: User
}
