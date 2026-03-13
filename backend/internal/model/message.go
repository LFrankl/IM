package model

import "time"

type Message struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromID     int64     `gorm:"not null;index" json:"from_id"`
	ToID       int64     `gorm:"not null;index:idx_chat" json:"to_id"`
	ChatType   string    `gorm:"not null;index:idx_chat" json:"chat_type"` // private/group
	MsgType    string    `gorm:"not null" json:"msg_type"`                 // text/image/file
	Content    string    `gorm:"not null" json:"content"`                  // JSON
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	IsRecalled bool      `gorm:"default:false" json:"is_recalled"`
	CreatedAt  time.Time `gorm:"index:idx_chat" json:"created_at"`

	From *User `gorm:"foreignKey:FromID" json:"from,omitempty"`
}
