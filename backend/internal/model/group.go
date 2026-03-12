package model

import "time"

type Group struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null;size:100" json:"name"`
	Avatar    string    `gorm:"default:''" json:"avatar"`
	Notice    string    `gorm:"default:''" json:"notice"`
	OwnerID   int64     `gorm:"not null" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Owner   *User         `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []GroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
}

type GroupMember struct {
	ID       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID  int64     `gorm:"not null;uniqueIndex:uniq_group_member" json:"group_id"`
	UserID   int64     `gorm:"not null;uniqueIndex:uniq_group_member" json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
