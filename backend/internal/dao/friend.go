package dao

import (
	"errors"
	"im-backend/internal/model"
	"time"

	"gorm.io/gorm"
)

type FriendDAO struct {
	db *gorm.DB
}

func NewFriendDAO(db *gorm.DB) *FriendDAO {
	return &FriendDAO{db: db}
}

// ListFriends 返回用户的所有好友（含好友信息）
func (d *FriendDAO) ListFriends(userID int64) ([]model.Friendship, error) {
	var list []model.Friendship
	err := d.db.Preload("Friend").
		Where("user_id = ?", userID).
		Order("group_name, created_at").
		Find(&list).Error
	return list, err
}

// IsFriend 检查是否互为好友
func (d *FriendDAO) IsFriend(userID, friendID int64) (bool, error) {
	var count int64
	err := d.db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error
	return count > 0, err
}

// AddFriendship 双向添加好友关系
func (d *FriendDAO) AddFriendship(userID, friendID int64) error {
	now := time.Now()
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Friendship{
			UserID: userID, FriendID: friendID, CreatedAt: now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&model.Friendship{
			UserID: friendID, FriendID: userID, CreatedAt: now,
		}).Error
	})
}

// DeleteFriendship 双向删除好友关系
func (d *FriendDAO) DeleteFriendship(userID, friendID int64) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).
			Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ? AND friend_id = ?", friendID, userID).
			Delete(&model.Friendship{}).Error
	})
}

// UpdateRemark 修改好友备注
func (d *FriendDAO) UpdateRemark(userID, friendID int64, remark string) error {
	return d.db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("remark", remark).Error
}

// UpdateGroupName 修改好友分组
func (d *FriendDAO) UpdateGroupName(userID, friendID int64, groupName string) error {
	return d.db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("group_name", groupName).Error
}

// ListFriendIDs 返回用户的所有好友 ID（轻量查询，用于在线状态广播）
func (d *FriendDAO) ListFriendIDs(userID int64) ([]int64, error) {
	var ids []int64
	err := d.db.Model(&model.Friendship{}).
		Where("user_id = ?", userID).
		Pluck("friend_id", &ids).Error
	return ids, err
}

// --- 好友申请 ---

// CreateRequest 发送好友申请（若已有 pending 则复用）
func (d *FriendDAO) CreateRequest(fromID, toID int64, message string) (*model.FriendRequest, error) {
	// 检查是否已有 pending 申请
	var existing model.FriendRequest
	err := d.db.Where("from_id = ? AND to_id = ? AND status = 'pending'", fromID, toID).
		First(&existing).Error
	if err == nil {
		// 已有 pending，更新 message
		existing.Message = message
		existing.UpdatedAt = time.Now()
		return &existing, d.db.Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	req := &model.FriendRequest{
		FromID:  fromID,
		ToID:    toID,
		Message: message,
		Status:  "pending",
	}
	return req, d.db.Create(req).Error
}

// GetRequestByID 获取申请详情
func (d *FriendDAO) GetRequestByID(id int64) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := d.db.Preload("From").First(&req, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &req, err
}

// ListPendingReceived 收到的待处理申请
func (d *FriendDAO) ListPendingReceived(toID int64) ([]model.FriendRequest, error) {
	var list []model.FriendRequest
	err := d.db.Preload("From").
		Where("to_id = ? AND status = 'pending'", toID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// UpdateRequestStatus 更新申请状态
func (d *FriendDAO) UpdateRequestStatus(id int64, status string) error {
	return d.db.Model(&model.FriendRequest{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()}).Error
}

// HasPendingRequest 检查是否已存在申请
func (d *FriendDAO) HasPendingRequest(fromID, toID int64) (bool, error) {
	var count int64
	err := d.db.Model(&model.FriendRequest{}).
		Where("from_id = ? AND to_id = ? AND status = 'pending'", fromID, toID).
		Count(&count).Error
	return count > 0, err
}

// CountPendingReceived 待处理申请数量（用于角标）
func (d *FriendDAO) CountPendingReceived(toID int64) (int64, error) {
	var count int64
	err := d.db.Model(&model.FriendRequest{}).
		Where("to_id = ? AND status = 'pending'", toID).
		Count(&count).Error
	return count, err
}
