package dao

import (
	"im-backend/internal/model"

	"gorm.io/gorm"
)

type GroupDAO struct {
	db *gorm.DB
}

func NewGroupDAO(db *gorm.DB) *GroupDAO {
	return &GroupDAO{db: db}
}

func (d *GroupDAO) Create(g *model.Group) error {
	return d.db.Create(g).Error
}

func (d *GroupDAO) GetByID(id int64) (*model.Group, error) {
	var g model.Group
	err := d.db.Preload("Owner").First(&g, id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (d *GroupDAO) ListByUserID(userID int64) ([]model.Group, error) {
	var groups []model.Group
	err := d.db.Preload("Owner").
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Order("groups.updated_at DESC").
		Find(&groups).Error
	return groups, err
}

func (d *GroupDAO) Search(keyword string) ([]model.Group, error) {
	var groups []model.Group
	err := d.db.Preload("Owner").
		Where("name LIKE ?", "%"+keyword+"%").
		Limit(20).
		Find(&groups).Error
	return groups, err
}

func (d *GroupDAO) Delete(id int64) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&model.GroupMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Group{}, id).Error
	})
}

// Members
func (d *GroupDAO) GetMembers(groupID int64) ([]model.GroupMember, error) {
	var members []model.GroupMember
	err := d.db.Preload("User").
		Where("group_id = ?", groupID).
		Find(&members).Error
	return members, err
}

func (d *GroupDAO) IsMember(groupID, userID int64) (bool, error) {
	var count int64
	err := d.db.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	return count > 0, err
}

func (d *GroupDAO) AddMember(groupID, userID int64) error {
	m := &model.GroupMember{GroupID: groupID, UserID: userID}
	return d.db.Create(m).Error
}

func (d *GroupDAO) RemoveMember(groupID, userID int64) error {
	return d.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupMember{}).Error
}

func (d *GroupDAO) MemberCount(groupID int64) (int64, error) {
	var count int64
	err := d.db.Model(&model.GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return count, err
}

func (d *GroupDAO) UpdateSettings(groupID int64, allowInvite bool) error {
	return d.db.Model(&model.Group{}).Where("id = ?", groupID).Update("allow_invite", allowInvite).Error
}

// GroupInvite DAO

func (d *GroupDAO) CreateInvite(inv *model.GroupInvite) error {
	return d.db.Create(inv).Error
}

func (d *GroupDAO) GetInvite(id int64) (*model.GroupInvite, error) {
	var inv model.GroupInvite
	err := d.db.Preload("Group").Preload("Inviter").First(&inv, id).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (d *GroupDAO) GetPendingInvite(groupID, inviteeID int64) (*model.GroupInvite, error) {
	var inv model.GroupInvite
	err := d.db.Where("group_id = ? AND invitee_id = ? AND status = 'pending'", groupID, inviteeID).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (d *GroupDAO) UpdateInviteStatus(id int64, status string) error {
	return d.db.Model(&model.GroupInvite{}).Where("id = ?", id).Update("status", status).Error
}

func (d *GroupDAO) ListPendingInvitesForUser(userID int64) ([]model.GroupInvite, error) {
	var invs []model.GroupInvite
	err := d.db.Preload("Group").Preload("Inviter").
		Where("invitee_id = ? AND status = 'pending'", userID).
		Order("created_at DESC").
		Find(&invs).Error
	return invs, err
}

// Group messages
func (d *GroupDAO) ListGroupConversations(userID int64) ([]int64, error) {
	var groupIDs []int64
	err := d.db.Model(&model.GroupMember{}).
		Where("user_id = ?", userID).
		Pluck("group_id", &groupIDs).Error
	return groupIDs, err
}
