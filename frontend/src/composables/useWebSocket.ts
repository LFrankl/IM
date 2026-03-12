import { ref, onUnmounted } from 'vue'
import type { WSMessage } from '@/types/chat'

type MessageHandler = (msg: WSMessage) => void

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
const RECONNECT_DELAY = 3000
const MAX_RECONNECT = 5

export function useWebSocket() {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  let reconnectCount = 0
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  const handlers: MessageHandler[] = []

  function connect(token: string) {
    if (ws.value?.readyState === WebSocket.OPEN) return

    ws.value = new WebSocket(`${WS_URL}?token=${token}`)

    ws.value.onopen = () => {
      connected.value = true
      reconnectCount = 0
    }

    ws.value.onclose = () => {
      connected.value = false
      tryReconnect(token)
    }

    ws.value.onerror = () => {
      ws.value?.close()
    }

    ws.value.onmessage = (e) => {
      try {
        const msg: WSMessage = JSON.parse(e.data)
        handlers.forEach((h) => h(msg))
      } catch {
        // 忽略非 JSON 消息
      }
    }
  }

  function tryReconnect(token: string) {
    if (reconnectCount >= MAX_RECONNECT) return
    reconnectCount++
    reconnectTimer = setTimeout(() => connect(token), RECONNECT_DELAY)
  }

  function send(data: object) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(data))
    }
  }

  function onMessage(handler: MessageHandler) {
    handlers.push(handler)
    return () => {
      const idx = handlers.indexOf(handler)
      if (idx !== -1) handlers.splice(idx, 1)
    }
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    reconnectCount = MAX_RECONNECT // 阻止重连
    ws.value?.close()
    ws.value = null
    connected.value = false
  }

  onUnmounted(disconnect)

  return { connected, connect, send, onMessage, disconnect }
}
