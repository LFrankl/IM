package handler

import (
	"errors"
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

type GroupHandler struct {
	groupSvc *service.GroupService
}

func NewGroupHandler(groupSvc *service.GroupService) *GroupHandler {
	return &GroupHandler{groupSvc: groupSvc}
}

func (h *GroupHandler) ListMyGroups(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.groupSvc.ListMyGroups(userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, list)
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var body struct {
		Name      string  `json:"name"`
		MemberIDs []int64 `json:"member_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		response.BadRequest(c, "群名不能为空")
		return
	}
	userID := middleware.GetUserID(c)
	g, err := h.groupSvc.CreateGroup(userID, body.Name, body.MemberIDs)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, g)
}

func (h *GroupHandler) SearchGroups(c *gin.Context) {
	kw := c.Query("q")
	if kw == "" {
		response.BadRequest(c, "请输入关键词")
		return
	}
	list, err := h.groupSvc.SearchGroups(kw)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, list)
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	userID := middleware.GetUserID(c)
	g, members, err := h.groupSvc.GetGroup(userID, groupID)
	if err != nil {
		if errors.Is(err, service.ErrNotGroupMember) {
			response.Forbidden(c)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OK(c, gin.H{"group": g, "members": members})
}

func (h *GroupHandler) GetMembers(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	userID := middleware.GetUserID(c)
	_, members, err := h.groupSvc.GetGroup(userID, groupID)
	if err != nil {
		response.Forbidden(c)
		return
	}
	response.OK(c, members)
}

func (h *GroupHandler) JoinGroup(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.JoinGroup(userID, groupID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.LeaveGroup(userID, groupID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *GroupHandler) KickMember(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	targetID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.KickMember(userID, groupID, targetID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *GroupHandler) DisbandGroup(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.DisbandGroup(userID, groupID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *GroupHandler) UpdateSettings(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	var body struct {
		AllowInvite bool `json:"allow_invite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.UpdateSettings(userID, groupID, body.AllowInvite); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *GroupHandler) InviteMember(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	var body struct {
		InviteeID int64 `json:"invitee_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.InviteeID == 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	inv, err := h.groupSvc.InviteToGroup(userID, groupID, body.InviteeID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, inv)
}

func (h *GroupHandler) HandleInvite(c *gin.Context) {
	inviteID, err := strconv.ParseInt(c.Param("inviteId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的邀请ID")
		return
	}
	var body struct {
		Accept bool `json:"accept"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.HandleInvite(userID, inviteID, body.Accept); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *GroupHandler) UpdateAvatar(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "请选择图片文件")
		return
	}
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

	if err := os.MkdirAll("./data/uploads/group-avatars", 0755); err != nil {
		response.InternalError(c)
		return
	}

	userID := middleware.GetUserID(c)
	filename := fmt.Sprintf("group_%d_%d%s", groupID, time.Now().UnixMilli(), ext)
	savePath := filepath.Join("./data/uploads/group-avatars", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.InternalError(c)
		return
	}

	avatarURL := "/uploads/group-avatars/" + filename
	oldAvatar, err := h.groupSvc.UpdateGroupAvatar(userID, groupID, avatarURL)
	if err != nil {
		os.Remove(savePath)
		response.BadRequest(c, err.Error())
		return
	}

	// 删除旧头像文件
	if oldAvatar != "" {
		os.Remove("." + oldAvatar)
	}

	response.OK(c, gin.H{"avatar": avatarURL})
}

func (h *GroupHandler) ListMyInvites(c *gin.Context) {
	userID := middleware.GetUserID(c)
	invites, err := h.groupSvc.ListMyInvites(userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, invites)
}

func (h *GroupHandler) GetMessages(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的群组ID")
		return
	}
	beforeID, _ := strconv.ParseInt(c.Query("before_id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit"))
	userID := middleware.GetUserID(c)
	msgs, err := h.groupSvc.GetGroupHistory(userID, groupID, beforeID, limit)
	if err != nil {
		if errors.Is(err, service.ErrNotGroupMember) {
			response.Forbidden(c)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OK(c, msgs)
}

func (h *GroupHandler) RecallMessage(c *gin.Context) {
	msgID, err := strconv.ParseInt(c.Param("msgId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的消息ID")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.groupSvc.RecallGroupMessage(userID, msgID); err != nil {
		switch err {
		case service.ErrRecallTimeout:
			response.Fail(c, 400, err.Error())
		case service.ErrRecallForbidden:
			response.Fail(c, 403, err.Error())
		case service.ErrMsgNotFound:
			response.Fail(c, 404, err.Error())
		default:
			response.InternalError(c)
		}
		return
	}
	response.OK(c, nil)
}
