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
	mu            sync.RWMutex
	clients       map[int64]*Client // userID -> client
	friendsLoader func(userID int64) []int64
}

var Global = &Hub{
	clients: make(map[int64]*Client),
}

// SetFriendsLoader 注入好友 ID 加载函数，避免循环依赖
func (h *Hub) SetFriendsLoader(fn func(userID int64) []int64) {
	h.friendsLoader = fn
}

func (h *Hub) Register(userID int64, client *Client) {
	h.mu.Lock()
	// 踢掉旧连接
	if old, ok := h.clients[userID]; ok {
		close(old.send)
	}
	h.clients[userID] = client
	h.mu.Unlock()

	// 广播上线给在线好友
	h.broadcastPresence(userID, "friend_online")
}

func (h *Hub) Unregister(userID int64) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()

	// 广播下线给在线好友
	h.broadcastPresence(userID, "friend_offline")
}

// broadcastPresence 向该用户的所有在线好友推送上/下线事件
func (h *Hub) broadcastPresence(userID int64, msgType string) {
	if h.friendsLoader == nil {
		return
	}
	friendIDs := h.friendsLoader(userID)
	payload := map[string]int64{"user_id": userID}
	for _, fid := range friendIDs {
		h.SendToUser(fid, msgType, payload)
	}
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
