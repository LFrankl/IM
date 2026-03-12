package service

import (
	"encoding/json"
	"errors"
	"im-backend/internal/dao"
	"im-backend/internal/model"
	"time"
)

type SpaceService struct {
	spaceDAO  *dao.SpaceDAO
	friendDAO *dao.FriendDAO
	userDAO   *dao.UserDAO
}

func NewSpaceService(spaceDAO *dao.SpaceDAO, friendDAO *dao.FriendDAO, userDAO *dao.UserDAO) *SpaceService {
	return &SpaceService{spaceDAO: spaceDAO, friendDAO: friendDAO, userDAO: userDAO}
}

// CreatePost 发帖
func (s *SpaceService) CreatePost(userID int64, content string, images []string) (*model.SpacePost, error) {
	if content == "" && len(images) == 0 {
		return nil, errors.New("内容不能为空")
	}
	imagesJSON, _ := json.Marshal(images)
	post := &model.SpacePost{
		UserID:    userID,
		Content:   content,
		Images:    string(imagesJSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.spaceDAO.CreatePost(post); err != nil {
		return nil, err
	}
	// 加载用户信息
	user, _ := s.userDAO.GetByID(userID)
	post.User = user
	return post, nil
}

// GetUserPosts 获取某用户的帖子（带当前访问者的点赞状态）
func (s *SpaceService) GetUserPosts(viewerID, targetID int64, beforeID int64) ([]model.SpacePost, error) {
	posts, err := s.spaceDAO.ListByUser(targetID, beforeID, 20)
	if err != nil {
		return nil, err
	}
	return s.fillLiked(viewerID, posts)
}

// GetFeed 好友动态
func (s *SpaceService) GetFeed(userID int64, beforeID int64) ([]model.SpacePost, error) {
	friends, err := s.friendDAO.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	friendIDs := make([]int64, 0, len(friends))
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendID)
	}
	posts, err := s.spaceDAO.Feed(userID, friendIDs, beforeID, 20)
	if err != nil {
		return nil, err
	}
	return s.fillLiked(userID, posts)
}

// DeletePost 删帖
func (s *SpaceService) DeletePost(userID, postID int64) error {
	return s.spaceDAO.DeletePost(postID, userID)
}

// LikePost 点赞
func (s *SpaceService) LikePost(userID, postID int64) error {
	liked, err := s.spaceDAO.IsLiked(postID, userID)
	if err != nil {
		return err
	}
	if liked {
		return errors.New("已经点赞")
	}
	return s.spaceDAO.Like(postID, userID)
}

// UnlikePost 取消点赞
func (s *SpaceService) UnlikePost(userID, postID int64) error {
	return s.spaceDAO.Unlike(postID, userID)
}

// AddComment 发评论
func (s *SpaceService) AddComment(userID, postID int64, content string) (*model.SpaceComment, error) {
	if content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	comment := &model.SpaceComment{
		PostID:    postID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := s.spaceDAO.AddComment(comment); err != nil {
		return nil, err
	}
	user, _ := s.userDAO.GetByID(userID)
	comment.User = user
	return comment, nil
}

// DeleteComment 删评论
func (s *SpaceService) DeleteComment(userID, commentID int64) error {
	return s.spaceDAO.DeleteComment(commentID, userID)
}

func (s *SpaceService) fillLiked(viewerID int64, posts []model.SpacePost) ([]model.SpacePost, error) {
	if len(posts) == 0 {
		return posts, nil
	}
	ids := make([]int64, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	likedMap, err := s.spaceDAO.LikedPostIDs(viewerID, ids)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Liked = likedMap[posts[i].ID]
	}
	return posts, nil
}
