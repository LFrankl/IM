package handler

import (
	"errors"
	"im-backend/internal/middleware"
	"im-backend/internal/service"
	"im-backend/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendSvc *service.FriendService
}

func NewFriendHandler(friendSvc *service.FriendService) *FriendHandler {
	return &FriendHandler{friendSvc: friendSvc}
}

func (h *FriendHandler) ListFriends(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.friendSvc.ListFriends(userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, list)
}

func (h *FriendHandler) SendRequest(c *gin.Context) {
	var req struct {
		ToID    int64  `json:"to_id" binding:"required"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	result, err := h.friendSvc.SendRequest(userID, req.ToID, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyFriend):
			response.Fail(c, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrRequestSelf),
			errors.Is(err, service.ErrUserNotFound):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c)
		}
		return
	}
	response.OK(c, result)
}

func (h *FriendHandler) ListRequests(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.friendSvc.ListPendingRequests(userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, list)
}

func (h *FriendHandler) HandleRequest(c *gin.Context) {
	reqID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的申请ID")
		return
	}
	var body struct {
		Action string `json:"action" binding:"required"` // accept / reject
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.Action != "accept" && body.Action != "reject" {
		response.BadRequest(c, "action 必须为 accept 或 reject")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.friendSvc.HandleRequest(userID, reqID, body.Action == "accept"); err != nil {
		switch {
		case errors.Is(err, service.ErrRequestNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrRequestNotYours):
			response.Forbidden(c)
		case errors.Is(err, service.ErrRequestNotPending):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c)
		}
		return
	}
	response.OK(c, nil)
}

func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.friendSvc.DeleteFriend(userID, friendID); err != nil {
		if errors.Is(err, service.ErrNotFriend) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}

func (h *FriendHandler) UpdateRemark(c *gin.Context) {
	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	var body struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.friendSvc.UpdateRemark(userID, friendID, body.Remark); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}

func (h *FriendHandler) UpdateGroup(c *gin.Context) {
	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	var body struct {
		GroupName string `json:"group_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.friendSvc.UpdateGroup(userID, friendID, body.GroupName); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, nil)
}

func (h *FriendHandler) CountPending(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.friendSvc.CountPendingRequests(userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"count": count})
}
