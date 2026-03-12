/**
 * 全局单例 WebSocket —— 整个 App 共用一个连接
 * 在 MainView 挂载时 connect，卸载时 disconnect
 */
import { ref } from 'vue'
import type { WSMessage } from '@/types/chat'

type MessageHandler = (msg: WSMessage) => void

const WS_BASE = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
const RECONNECT_DELAY = 3000
const MAX_RECONNECT = 10

let ws: WebSocket | null = null
let reconnectCount = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let currentToken = ''
const handlers = new Set<MessageHandler>()

export const wsConnected = ref(false)

function connect(token: string) {
  currentToken = token
  if (ws?.readyState === WebSocket.OPEN) return

  ws = new WebSocket(`${WS_BASE}?token=${token}`)

  ws.onopen = () => {
    console.log('[WS] connected')
    wsConnected.value = true
    reconnectCount = 0
  }

  ws.onclose = (e) => {
    console.warn('[WS] closed, code=', e.code, 'reason=', e.reason)
    wsConnected.value = false
    ws = null
    if (reconnectCount < MAX_RECONNECT) {
      reconnectCount++
      reconnectTimer = setTimeout(() => connect(currentToken), RECONNECT_DELAY)
    }
  }

  ws.onerror = (e) => {
    console.error('[WS] error', e)
    ws?.close()
  }

  ws.onmessage = (e) => {
    console.log('[WS] received:', e.data)
    try {
      const msg: WSMessage = JSON.parse(e.data)
      handlers.forEach((h) => h(msg))
    } catch { /* ignore */ }
  }
}

function disconnect() {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  reconnectCount = MAX_RECONNECT
  ws?.close()
  ws = null
  wsConnected.value = false
}

function send(data: object) {
  console.log('[WS] send, readyState=', ws?.readyState, 'data=', data)
  if (ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(data))
  } else {
    console.warn('[WS] not connected, message dropped')
  }
}

function onMessage(handler: MessageHandler): () => void {
  handlers.add(handler)
  return () => handlers.delete(handler)
}

export function useWS() {
  return { connect, disconnect, send, onMessage, connected: wsConnected }
}
