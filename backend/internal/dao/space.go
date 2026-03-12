package dao

import (
	"im-backend/internal/model"

	"gorm.io/gorm"
)

type SpaceDAO struct {
	db *gorm.DB
}

func NewSpaceDAO(db *gorm.DB) *SpaceDAO {
	return &SpaceDAO{db: db}
}

// CreatePost 发帖
func (d *SpaceDAO) CreatePost(post *model.SpacePost) error {
	return d.db.Create(post).Error
}

// GetPost 获取单篇帖子（含评论和作者）
func (d *SpaceDAO) GetPost(id int64) (*model.SpacePost, error) {
	var post model.SpacePost
	err := d.db.Preload("User").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("id ASC")
		}).
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// ListByUser 某用户的帖子列表（分页）
func (d *SpaceDAO) ListByUser(userID int64, beforeID int64, limit int) ([]model.SpacePost, error) {
	q := d.db.Preload("User").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("id ASC")
		}).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	var posts []model.SpacePost
	return posts, q.Find(&posts).Error
}

// Feed 好友动态（自己+好友的帖子）
func (d *SpaceDAO) Feed(userID int64, friendIDs []int64, beforeID int64, limit int) ([]model.SpacePost, error) {
	ids := append([]int64{userID}, friendIDs...)
	q := d.db.Preload("User").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("id ASC")
		}).
		Where("user_id IN ?", ids).
		Order("id DESC").
		Limit(limit)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	var posts []model.SpacePost
	return posts, q.Find(&posts).Error
}

// DeletePost 删帖（仅作者）
func (d *SpaceDAO) DeletePost(postID, userID int64) error {
	return d.db.Where("id = ? AND user_id = ?", postID, userID).Delete(&model.SpacePost{}).Error
}

// IsLiked 是否已点赞
func (d *SpaceDAO) IsLiked(postID, userID int64) (bool, error) {
	var count int64
	err := d.db.Model(&model.SpaceLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}

// LikedPostIDs 批量查哪些帖子已点赞
func (d *SpaceDAO) LikedPostIDs(userID int64, postIDs []int64) (map[int64]bool, error) {
	var likes []model.SpaceLike
	err := d.db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&likes).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int64]bool, len(likes))
	for _, l := range likes {
		m[l.PostID] = true
	}
	return m, nil
}

// Like 点赞
func (d *SpaceDAO) Like(postID, userID int64) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		like := &model.SpaceLike{PostID: postID, UserID: userID}
		if err := tx.Create(like).Error; err != nil {
			return err
		}
		return tx.Model(&model.SpacePost{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

// Unlike 取消点赞
func (d *SpaceDAO) Unlike(postID, userID int64) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&model.SpaceLike{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&model.SpacePost{}).Where("id = ? AND like_count > 0", postID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
	})
}

// AddComment 发评论
func (d *SpaceDAO) AddComment(comment *model.SpaceComment) error {
	return d.db.Create(comment).Error
}

// DeleteComment 删评论（仅作者）
func (d *SpaceDAO) DeleteComment(commentID, userID int64) error {
	return d.db.Where("id = ? AND user_id = ?", commentID, userID).Delete(&model.SpaceComment{}).Error
}
