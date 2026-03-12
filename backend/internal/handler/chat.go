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

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.chatSvc.ListConversations(userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, list)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	beforeID, _ := strconv.ParseInt(c.Query("before_id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit"))

	userID := middleware.GetUserID(c)
	msgs, err := h.chatSvc.GetHistory(userID, targetID, beforeID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, msgs)
}

func (h *ChatHandler) MarkRead(c *gin.Context) {
	fromID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.chatSvc.MarkRead(userID, fromID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}

func (h *ChatHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件不存在")
		return
	}

	// 限制大小 20MB
	if file.Size > 20*1024*1024 {
		response.BadRequest(c, "文件过大，最大 20MB")
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dir := "./data/uploads"
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.InternalError(c)
		return
	}
	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"url": "/uploads/" + filename})
}
