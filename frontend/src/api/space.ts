import client from './client'

export interface SpacePost {
  id: number
  user_id: number
  content: string
  images: string   // JSON string []
  like_count: number
  liked: boolean
  created_at: string
  user?: { id: number; nickname: string; avatar: string }
  comments?: SpaceComment[]
}

export interface SpaceComment {
  id: number
  post_id: number
  user_id: number
  content: string
  created_at: string
  user?: { id: number; nickname: string; avatar: string }
}

export const spaceApi = {
  getFeed(beforeId?: number) {
    return client.get<SpacePost[]>('/api/space/feed', { params: { before_id: beforeId || undefined } })
  },
  getUserPosts(userId: number, beforeId?: number) {
    return client.get<SpacePost[]>(`/api/space/users/${userId}/posts`, { params: { before_id: beforeId || undefined } })
  },
  createPost(content: string, images: string[] = []) {
    return client.post<SpacePost>('/api/space/posts', { content, images })
  },
  deletePost(id: number) {
    return client.delete(`/api/space/posts/${id}`)
  },
  likePost(id: number) {
    return client.post(`/api/space/posts/${id}/like`)
  },
  unlikePost(id: number) {
    return client.delete(`/api/space/posts/${id}/like`)
  },
  addComment(postId: number, content: string) {
    return client.post<SpaceComment>(`/api/space/posts/${postId}/comments`, { content })
  },
  deleteComment(id: number) {
    return client.delete(`/api/space/comments/${id}`)
  },
}
