package model

import "time"

type SpacePost struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Content   string    `gorm:"not null" json:"content"`
	Images    string    `gorm:"default:'[]'" json:"images"` // JSON 数组
	LikeCount int       `gorm:"default:0" json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User     *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Comments []SpaceComment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
	Liked    bool           `gorm:"-" json:"liked"` // 当前用户是否点赞（非DB字段）
}

type SpaceComment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    int64     `gorm:"not null;index" json:"post_id"`
	UserID    int64     `gorm:"not null" json:"user_id"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type SpaceLike struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    int64     `gorm:"not null;uniqueIndex:uniq_like" json:"post_id"`
	UserID    int64     `gorm:"not null;uniqueIndex:uniq_like" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
