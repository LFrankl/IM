package handler

import (
	"encoding/json"
	"im-backend/internal/middleware"
	"im-backend/internal/service"
	"im-backend/internal/ws"
	"im-backend/pkg/response"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSHandler struct {
	chatSvc  *service.ChatService
	groupSvc *service.GroupService
	hub      *ws.Hub
}

func NewWSHandler(chatSvc *service.ChatService, groupSvc *service.GroupService, hub *ws.Hub) *WSHandler {
	return &WSHandler{chatSvc: chatSvc, groupSvc: groupSvc, hub: hub}
}

// wsIncoming 客户端发来的消息结构
type wsIncoming struct {
	Type    string          `json:"type"`
	ToID    int64           `json:"to_id"`    // 私聊：对方用户ID；群聊：群ID
	MsgType string          `json:"msg_type"`
	Content json.RawMessage `json:"content"`
}

func (h *WSHandler) Handle(c *gin.Context) {
	userID := middleware.GetUserID(c)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.BadRequest(c, "websocket upgrade failed")
		return
	}

	client := ws.NewClient(userID, conn, h.hub, h.onMessage)
	h.hub.Register(userID, client)
	client.Run() // 阻塞直到断开
}

func (h *WSHandler) onMessage(userID int64, raw []byte) {
	var msg wsIncoming
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("ws invalid json from %d: %v", userID, err)
		return
	}

	switch msg.Type {
	case "chat_private":
		h.handleChatPrivate(userID, msg)
	case "chat_group":
		h.handleChatGroup(userID, msg)
	case "heartbeat":
		h.hub.SendToUser(userID, "heartbeat_ack", nil)
	}
}

func (h *WSHandler) handleChatPrivate(fromID int64, msg wsIncoming) {
	if msg.ToID == 0 || msg.MsgType == "" || len(msg.Content) == 0 {
		return
	}

	var content any
	if err := json.Unmarshal(msg.Content, &content); err != nil {
		return
	}

	saved, err := h.chatSvc.SendPrivate(service.SendMsgInput{
		FromID:  fromID,
		ToID:    msg.ToID,
		MsgType: msg.MsgType,
		Content: content,
	})
	if err != nil {
		log.Printf("ws send private from %d to %d: %v", fromID, msg.ToID, err)
		// 通知发送方失败
		h.hub.SendToUser(fromID, "error", map[string]string{"message": err.Error()})
		return
	}

	// 回传给发送方（确认+完整消息体含 ID）
	h.hub.SendToUser(fromID, "message_sent", saved)
}

func (h *WSHandler) handleChatGroup(fromID int64, msg wsIncoming) {
	if msg.ToID == 0 || msg.MsgType == "" || len(msg.Content) == 0 {
		return
	}
	var content any
	if err := json.Unmarshal(msg.Content, &content); err != nil {
		return
	}
	saved, err := h.groupSvc.SendGroupMessage(fromID, msg.ToID, msg.MsgType, content)
	if err != nil {
		log.Printf("ws send group from %d to group %d: %v", fromID, msg.ToID, err)
		h.hub.SendToUser(fromID, "error", map[string]string{"message": err.Error()})
		return
	}
	// 回传给发送方确认
	h.hub.SendToUser(fromID, "message_sent", saved)
}
