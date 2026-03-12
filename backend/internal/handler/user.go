package handler

import (
	"im-backend/internal/middleware"
	"im-backend/internal/service"
	"im-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.OK(c, []any{})
		return
	}
	userID := middleware.GetUserID(c)
	users, err := h.userSvc.Search(q, userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	user, err := h.userSvc.GetByID(id)
	if err != nil || user == nil {
		response.NotFound(c, "用户不存在")
		return
	}
	response.OK(c, user)
}
