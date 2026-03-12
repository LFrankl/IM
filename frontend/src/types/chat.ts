export type MsgType = 'text' | 'image' | 'file'
export type ChatType = 'private' | 'group'

export interface TextContent {
  text: string
}

export interface ImageContent {
  url: string
  width?: number
  height?: number
}

export interface FileContent {
  url: string
  name: string
  size: number
}

export interface Message {
  id: number
  from_id: number
  to_id: number
  chat_type: ChatType
  msg_type: MsgType
  content: TextContent | ImageContent | FileContent
  is_read: boolean
  created_at: string
  from?: {
    id: number
    nickname: string
    avatar: string
  }
}

export interface Conversation {
  id: string          // `${chat_type}:${to_id}`
  chat_type: ChatType
  target_id: number
  name: string
  avatar: string
  last_message: Message | null
  unread_count: number
  updated_at: string
}

// WebSocket 消息类型
export type WSMessageType =
  | 'message'
  | 'message_sent'
  | 'friend_request'
  | 'friend_accepted'
  | 'friend_online'
  | 'friend_offline'
  | 'error'
  | 'heartbeat_ack'

export interface WSMessage {
  type: WSMessageType
  data?: unknown
}

// 客户端发送的消息
export interface WSSendChat {
  type: 'chat_private' | 'chat_group'
  to_id?: number
  group_id?: number
  msg_type: MsgType
  content: TextContent | ImageContent | FileContent
}
