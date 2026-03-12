package model

import "time"

type Group struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"not null;size:100" json:"name"`
	Avatar      string    `gorm:"default:''" json:"avatar"`
	Notice      string    `gorm:"default:''" json:"notice"`
	OwnerID     int64     `gorm:"not null" json:"owner_id"`
	AllowInvite bool      `gorm:"default:true" json:"allow_invite"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

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

// GroupInvite 群邀请记录
type GroupInvite struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID    int64     `gorm:"not null;index" json:"group_id"`
	InviterID  int64     `gorm:"not null" json:"inviter_id"`
	InviteeID  int64     `gorm:"not null;index" json:"invitee_id"`
	Status     string    `gorm:"default:'pending';size:20" json:"status"` // pending/accepted/rejected
	CreatedAt  time.Time `json:"created_at"`

	Group   *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Inviter *User  `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
}
