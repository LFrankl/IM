package dao

import (
	"errors"
	"im-backend/internal/model"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) Create(user *model.User) error {
	return d.db.Create(user).Error
}

func (d *UserDAO) GetByID(id int64) (*model.User, error) {
	var u model.User
	err := d.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (d *UserDAO) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := d.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (d *UserDAO) Update(user *model.User) error {
	return d.db.Save(user).Error
}

func (d *UserDAO) UpdateStatus(id int64, status string) error {
	return d.db.Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}

func (d *UserDAO) Search(keyword string, excludeID int64) ([]model.User, error) {
	var users []model.User
	err := d.db.Where("(username LIKE ? OR nickname LIKE ?) AND id != ?",
		"%"+keyword+"%", "%"+keyword+"%", excludeID).
		Limit(20).Find(&users).Error
	return users, err
}
