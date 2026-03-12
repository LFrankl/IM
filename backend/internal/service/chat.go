package service

import (
	"encoding/json"
	"errors"
	"im-backend/internal/dao"
	"im-backend/internal/model"
	"im-backend/internal/ws"
	"strconv"
	"time"
)

type ChatService struct {
	msgDAO    *dao.MessageDAO
	friendDAO *dao.FriendDAO
	userDAO   *dao.UserDAO
	hub       *ws.Hub
}

func NewChatService(msgDAO *dao.MessageDAO, friendDAO *dao.FriendDAO, userDAO *dao.UserDAO, hub *ws.Hub) *ChatService {
	return &ChatService{msgDAO: msgDAO, friendDAO: friendDAO, userDAO: userDAO, hub: hub}
}

var ErrNotFriendChat = errors.New("请先添加对方为好友")

type SendMsgInput struct {
	FromID  int64
	ToID    int64
	MsgType string // text/image/file
	Content any    // 会被序列化为 JSON
}

// SendPrivate 发送私聊消息
func (s *ChatService) SendPrivate(input SendMsgInput) (*model.Message, error) {
	// 必须是好友才能发消息
	ok, err := s.friendDAO.IsFriend(input.FromID, input.ToID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFriendChat
	}

	contentJSON, err := json.Marshal(input.Content)
	if err != nil {
		return nil, err
	}

	msg := &model.Message{
		FromID:    input.FromID,
		ToID:      input.ToID,
		ChatType:  "private",
		MsgType:   input.MsgType,
		Content:   string(contentJSON),
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	if err := s.msgDAO.Save(msg); err != nil {
		return nil, err
	}

	// 加载发送者信息
	from, _ := s.userDAO.GetByID(input.FromID)
	if from != nil {
		msg.From = from
	}

	// 推送给接收方
	s.hub.SendToUser(input.ToID, "message", msg)

	return msg, nil
}

// GetHistory 获取私聊历史
func (s *ChatService) GetHistory(userID, targetID int64, beforeID int64, limit int) ([]model.Message, error) {
	if limit <= 0 || limit > 50 {
		limit = 30
	}
	return s.msgDAO.GetHistory(userID, targetID, beforeID, limit)
}

// MarkRead 标记已读
func (s *ChatService) MarkRead(userID, fromID int64) error {
	return s.msgDAO.MarkRead(fromID, userID)
}

// ConversationDTO 会话列表项
type ConversationDTO struct {
	ID          string       `json:"id"`
	ChatType    string       `json:"chat_type"`
	TargetID    int64        `json:"target_id"`
	Name        string       `json:"name"`
	Avatar      string       `json:"avatar"`
	LastMessage *model.Message `json:"last_message"`
	UnreadCount int64        `json:"unread_count"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ListConversations 获取会话列表
func (s *ChatService) ListConversations(userID int64) ([]ConversationDTO, error) {
	rows, err := s.msgDAO.ListConversations(userID)
	if err != nil {
		return nil, err
	}

	unreadMap, err := s.msgDAO.CountUnread(userID)
	if err != nil {
		return nil, err
	}

	result := make([]ConversationDTO, 0, len(rows))
	for _, row := range rows {
		peer, err := s.userDAO.GetByID(row.PeerID)
		if err != nil || peer == nil {
			continue
		}
		msg := row.LastMsg
		result = append(result, ConversationDTO{
			ID:          "private:" + strconv.FormatInt(row.PeerID, 10),
			ChatType:    "private",
			TargetID:    row.PeerID,
			Name:        peer.Nickname,
			Avatar:      peer.Avatar,
			LastMessage: &msg,
			UnreadCount: unreadMap[row.PeerID],
			UpdatedAt:   msg.CreatedAt,
		})
	}
	return result, nil
}
