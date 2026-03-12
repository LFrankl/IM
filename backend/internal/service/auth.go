package service

import (
	"errors"
	"im-backend/internal/dao"
	"im-backend/internal/model"
	pkgjwt "im-backend/pkg/jwt"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userDAO   *dao.UserDAO
	jwtExpire time.Duration
}

func NewAuthService(userDAO *dao.UserDAO, jwtExpire time.Duration) *AuthService {
	return &AuthService{userDAO: userDAO, jwtExpire: jwtExpire}
}

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

type LoginInput struct {
	Username string
	Password string
}

type AuthResult struct {
	Token string
	User  *model.User
}

var (
	ErrUsernameTaken   = errors.New("用户名已被使用")
	ErrUserNotFound    = errors.New("用户不存在")
	ErrWrongPassword   = errors.New("密码错误")
	ErrInvalidUsername = errors.New("用户名只能包含字母、数字和下划线，长度 3-20")
	ErrInvalidPassword = errors.New("密码长度至少 6 位")
	ErrInvalidNickname = errors.New("昵称长度 1-20 个字符")
)

func (s *AuthService) Register(input RegisterInput) (*AuthResult, error) {
	if !isValidUsername(input.Username) {
		return nil, ErrInvalidUsername
	}
	if len(input.Password) < 6 {
		return nil, ErrInvalidPassword
	}
	l := utf8.RuneCountInString(input.Nickname)
	if l < 1 || l > 20 {
		return nil, ErrInvalidNickname
	}

	existing, err := s.userDAO.GetByUsername(input.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: input.Username,
		Password: string(hash),
		Nickname: input.Nickname,
		Status:   "online",
	}
	if err := s.userDAO.Create(user); err != nil {
		return nil, err
	}

	token, err := pkgjwt.Generate(user.ID, user.Username, s.jwtExpire)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(input LoginInput) (*AuthResult, error) {
	user, err := s.userDAO.GetByUsername(input.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, ErrWrongPassword
	}

	// 更新在线状态
	_ = s.userDAO.UpdateStatus(user.ID, "online")
	user.Status = "online"

	token, err := pkgjwt.Generate(user.ID, user.Username, s.jwtExpire)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Logout(userID int64) error {
	return s.userDAO.UpdateStatus(userID, "offline")
}

func isValidUsername(s string) bool {
	if len(s) < 3 || len(s) > 20 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
