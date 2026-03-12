package service

import (
	"errors"
	"im-backend/internal/dao"
	"im-backend/internal/model"
	"os"
	"path/filepath"
	"strings"
)

type UserService struct {
	userDAO *dao.UserDAO
}

func NewUserService(userDAO *dao.UserDAO) *UserService {
	return &UserService{userDAO: userDAO}
}

func (s *UserService) GetByID(id int64) (*model.User, error) {
	return s.userDAO.GetByID(id)
}

func (s *UserService) Search(keyword string, excludeID int64) ([]model.User, error) {
	return s.userDAO.Search(keyword, excludeID)
}

func (s *UserService) UpdateAvatar(userID int64, avatarURL string) (*model.User, error) {
	user, err := s.userDAO.GetByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}
	user.Avatar = avatarURL
	if err := s.userDAO.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateCover(userID int64, coverURL string) (*model.User, error) {
	user, err := s.userDAO.GetByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}
	oldCover := user.Cover
	user.Cover = coverURL
	if err := s.userDAO.Update(user); err != nil {
		return nil, err
	}
	if oldCover != "" && strings.HasPrefix(oldCover, "/uploads/covers/") {
		oldPath := filepath.Join("./data", oldCover)
		os.Remove(oldPath)
	}
	return user, nil
}

func (s *UserService) UpdateProfile(userID int64, nickname, bio string) (*model.User, error) {
	if nickname == "" {
		return nil, errors.New("昵称不能为空")
	}
	user, err := s.userDAO.GetByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}
	user.Nickname = nickname
	user.Bio = bio
	if err := s.userDAO.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}
