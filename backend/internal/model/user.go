package model

import "time"

type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Nickname  string    `gorm:"not null;size:50" json:"nickname"`
	Avatar    string    `gorm:"default:''" json:"avatar"`
	Cover     string    `gorm:"default:''" json:"cover"`
	Bio       string    `gorm:"default:''" json:"bio"`
	Status    string    `gorm:"default:'offline'" json:"status"` // online/offline/busy
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Friendship struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:uniq_friendship" json:"user_id"`
	FriendID  int64     `gorm:"not null;uniqueIndex:uniq_friendship" json:"friend_id"`
	Remark    string    `gorm:"default:''" json:"remark"`
	GroupName string    `gorm:"default:'我的好友'" json:"group_name"`
	CreatedAt time.Time `json:"created_at"`

	Friend *User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}

type FriendRequest struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromID    int64     `gorm:"not null" json:"from_id"`
	ToID      int64     `gorm:"not null" json:"to_id"`
	Message   string    `gorm:"default:''" json:"message"`
	Status    string    `gorm:"default:'pending'" json:"status"` // pending/accepted/rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	From *User `gorm:"foreignKey:FromID" json:"from,omitempty"`
}
