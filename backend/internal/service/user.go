package service

import (
	"im-backend/internal/dao"
	"im-backend/internal/model"
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
