package handler

import (
	"errors"
	"im-backend/internal/middleware"
	"im-backend/internal/service"
	"im-backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
	userSvc *service.UserService
}

func NewAuthHandler(authSvc *service.AuthService, userSvc *service.UserService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, userSvc: userSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.authSvc.Register(service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			response.Fail(c, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrInvalidUsername),
			errors.Is(err, service.ErrInvalidPassword),
			errors.Is(err, service.ErrInvalidNickname):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, gin.H{"token": result.Token, "user": result.User})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.authSvc.Login(service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound),
			errors.Is(err, service.ErrWrongPassword):
			response.Fail(c, http.StatusUnauthorized, "用户名或密码错误")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, gin.H{"token": result.Token, "user": result.User})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	_ = h.authSvc.Logout(userID)
	response.OK(c, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.userSvc.GetByID(userID)
	if err != nil || user == nil {
		response.NotFound(c, "用户不存在")
		return
	}
	response.OK(c, user)
}
