package handler

import (
	"im-backend/internal/middleware"
	"im-backend/internal/service"
	"im-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SpaceHandler struct {
	spaceSvc *service.SpaceService
}

func NewSpaceHandler(spaceSvc *service.SpaceService) *SpaceHandler {
	return &SpaceHandler{spaceSvc: spaceSvc}
}

func (h *SpaceHandler) GetFeed(c *gin.Context) {
	userID := middleware.GetUserID(c)
	beforeID, _ := strconv.ParseInt(c.Query("before_id"), 10, 64)
	posts, err := h.spaceSvc.GetFeed(userID, beforeID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, posts)
}

func (h *SpaceHandler) GetUserPosts(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	userID := middleware.GetUserID(c)
	beforeID, _ := strconv.ParseInt(c.Query("before_id"), 10, 64)
	posts, err := h.spaceSvc.GetUserPosts(userID, targetID, beforeID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, posts)
}

func (h *SpaceHandler) CreatePost(c *gin.Context) {
	var body struct {
		Content string   `json:"content"`
		Images  []string `json:"images"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	post, err := h.spaceSvc.CreatePost(userID, body.Content, body.Images)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, post)
}

func (h *SpaceHandler) DeletePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的帖子ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.spaceSvc.DeletePost(userID, postID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}

func (h *SpaceHandler) LikePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的帖子ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.spaceSvc.LikePost(userID, postID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SpaceHandler) UnlikePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的帖子ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.spaceSvc.UnlikePost(userID, postID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}

func (h *SpaceHandler) AddComment(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的帖子ID")
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Content == "" {
		response.BadRequest(c, "评论内容不能为空")
		return
	}
	userID := middleware.GetUserID(c)
	comment, err := h.spaceSvc.AddComment(userID, postID, body.Content)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, comment)
}

func (h *SpaceHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.spaceSvc.DeleteComment(userID, commentID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}
