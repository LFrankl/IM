package dao

import (
	"encoding/json"
	"im-backend/internal/model"

	"gorm.io/gorm"
)

type MessageDAO struct {
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

// Save 保存一条消息，content 已是 JSON 字符串
func (d *MessageDAO) Save(msg *model.Message) error {
	return d.db.Create(msg).Error
}

// GetHistory 获取私聊历史（游标分页，越新越靠后）
func (d *MessageDAO) GetHistory(userID, targetID int64, beforeID int64, limit int) ([]model.Message, error) {
	q := d.db.Preload("From").
		Where("chat_type = 'private' AND ((from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?))",
			userID, targetID, targetID, userID).
		Order("id DESC").
		Limit(limit)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	var msgs []model.Message
	if err := q.Find(&msgs).Error; err != nil {
		return nil, err
	}
	// 翻转为正序
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// MarkRead 标记某会话消息已读
func (d *MessageDAO) MarkRead(fromID, toID int64) error {
	return d.db.Model(&model.Message{}).
		Where("chat_type = 'private' AND from_id = ? AND to_id = ? AND is_read = false", fromID, toID).
		Update("is_read", true).Error
}

// CountUnread 某用户收到的未读消息数（按发送方聚合）
func (d *MessageDAO) CountUnread(toID int64) (map[int64]int64, error) {
	type row struct {
		FromID int64
		Count  int64
	}
	var rows []row
	err := d.db.Model(&model.Message{}).
		Select("from_id, count(*) as count").
		Where("chat_type = 'private' AND to_id = ? AND is_read = false", toID).
		Group("from_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int64]int64, len(rows))
	for _, r := range rows {
		m[r.FromID] = r.Count
	}
	return m, nil
}

// ListConversations 会话列表：每个对话取最新一条消息
func (d *MessageDAO) ListConversations(userID int64) ([]ConversationRow, error) {
	// 取与该用户相关的所有私聊消息里，每个"对话对"的最新一条
	query := `
		SELECT
			CASE WHEN from_id = ? THEN to_id ELSE from_id END AS peer_id,
			MAX(id) AS last_msg_id
		FROM messages
		WHERE chat_type = 'private' AND (from_id = ? OR to_id = ?)
		GROUP BY peer_id
		ORDER BY last_msg_id DESC
	`
	type peerRow struct {
		PeerID    int64
		LastMsgID int64
	}
	var peers []peerRow
	if err := d.db.Raw(query, userID, userID, userID).Scan(&peers).Error; err != nil {
		return nil, err
	}

	result := make([]ConversationRow, 0, len(peers))
	for _, p := range peers {
		var msg model.Message
		if err := d.db.Preload("From").First(&msg, p.LastMsgID).Error; err != nil {
			continue
		}
		result = append(result, ConversationRow{
			PeerID:  p.PeerID,
			LastMsg: msg,
		})
	}
	return result, nil
}

// GetGroupHistory 获取群聊历史（游标分页）
func (d *MessageDAO) GetGroupHistory(groupID int64, beforeID int64, limit int) ([]model.Message, error) {
	q := d.db.Preload("From").
		Where("chat_type = 'group' AND to_id = ?", groupID).
		Order("id DESC").
		Limit(limit)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	var msgs []model.Message
	if err := q.Find(&msgs).Error; err != nil {
		return nil, err
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// LastGroupMessage 获取群的最新一条消息
func (d *MessageDAO) LastGroupMessage(groupID int64) (*model.Message, error) {
	var msg model.Message
	err := d.db.Preload("From").
		Where("chat_type = 'group' AND to_id = ?", groupID).
		Order("id DESC").
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

type ConversationRow struct {
	PeerID  int64
	LastMsg model.Message
}

// GetByID 获取单条消息
func (d *MessageDAO) GetByID(id int64) (*model.Message, error) {
	var msg model.Message
	err := d.db.First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Recall 标记消息已撤回
func (d *MessageDAO) Recall(id int64) error {
	return d.db.Model(&model.Message{}).Where("id = ?", id).
		Updates(map[string]any{"is_recalled": true, "content": `{"text":"[消息已撤回]"}`}).Error
}

// BuildTextContent 构建文本消息 JSON
func BuildTextContent(text string) string {
	b, _ := json.Marshal(map[string]string{"text": text})
	return string(b)
}
