package ws

import (
	"encoding/json"
	"sync"
)

// Message 是 WebSocket 收发的消息结构
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]*Client // userID -> client
}

var Global = &Hub{
	clients: make(map[int64]*Client),
}

func (h *Hub) Register(userID int64, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// 踢掉旧连接
	if old, ok := h.clients[userID]; ok {
		close(old.send)
	}
	h.clients[userID] = client
}

func (h *Hub) Unregister(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

// SendToUser 向指定用户发送消息（异步）
func (h *Hub) SendToUser(userID int64, msgType string, data any) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := Message{Type: msgType, Data: payload}
	raw, _ := json.Marshal(msg)

	select {
	case client.send <- raw:
	default:
		// 发送缓冲满，丢弃
	}
}

// IsOnline 检查用户是否在线
func (h *Hub) IsOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// OnlineUserIDs 返回所有在线用户 ID
func (h *Hub) OnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]int64, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}
