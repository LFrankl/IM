package handler

import (
	"fmt"
	"im-backend/internal/middleware"
	"im-backend/internal/service"
	"im-backend/pkg/response"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "请选择图片文件")
		return
	}

	// 限制 5MB
	if file.Size > 5<<20 {
		response.BadRequest(c, "图片不能超过 5MB")
		return
	}

	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		response.BadRequest(c, "仅支持 jpg/png/gif/webp 格式")
		return
	}

	if err := os.MkdirAll("./data/uploads/avatars", 0755); err != nil {
		response.InternalError(c)
		return
	}

	userID := middleware.GetUserID(c)
	filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().UnixMilli(), ext)
	savePath := filepath.Join("./data/uploads/avatars", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.InternalError(c)
		return
	}

	avatarURL := "/uploads/avatars/" + filename
	user, err := h.userSvc.UpdateAvatar(userID, avatarURL)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var body struct {
		Nickname string `json:"nickname"`
		Bio      string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	user, err := h.userSvc.UpdateProfile(userID, body.Nickname, body.Bio)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) UploadCover(c *gin.Context) {
	file, err := c.FormFile("cover")
	if err != nil {
		response.BadRequest(c, "请选择图片文件")
		return
	}
	if file.Size > 10<<20 {
		response.BadRequest(c, "图片不能超过 10MB")
		return
	}
	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		response.BadRequest(c, "仅支持 jpg/png/gif/webp 格式")
		return
	}
	if err := os.MkdirAll("./data/uploads/covers", 0755); err != nil {
		response.InternalError(c)
		return
	}
	userID := middleware.GetUserID(c)
	filename := fmt.Sprintf("cover_%d_%d%s", userID, time.Now().UnixMilli(), ext)
	savePath := filepath.Join("./data/uploads/covers", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.InternalError(c)
		return
	}
	coverURL := "/uploads/covers/" + filename
	user, err := h.userSvc.UpdateCover(userID, coverURL)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, user)
}
