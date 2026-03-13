package service

import (
	"encoding/json"
	"errors"
	"im-backend/internal/dao"
	"im-backend/internal/model"
	"im-backend/internal/ws"
	"time"
)

var (
	ErrNotGroupMember  = errors.New("你不是该群成员")
	ErrNotGroupOwner   = errors.New("只有群主才能执行此操作")
	ErrAlreadyMember   = errors.New("已经是群成员")
	ErrInviteNotFound  = errors.New("邀请不存在")
	ErrInviteNotPending = errors.New("邀请已处理")
	ErrInviteNotForYou = errors.New("该邀请不属于你")
)

type GroupService struct {
	groupDAO *dao.GroupDAO
	msgDAO   *dao.MessageDAO
	userDAO  *dao.UserDAO
	hub      *ws.Hub
}

func NewGroupService(groupDAO *dao.GroupDAO, msgDAO *dao.MessageDAO, userDAO *dao.UserDAO, hub *ws.Hub) *GroupService {
	return &GroupService{groupDAO: groupDAO, msgDAO: msgDAO, userDAO: userDAO, hub: hub}
}

func marshalContent(content interface{}) (string, error) {
	b, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CreateGroup 创建群组
func (s *GroupService) CreateGroup(ownerID int64, name string, memberIDs []int64) (*model.Group, error) {
	g := &model.Group{Name: name, OwnerID: ownerID}
	if err := s.groupDAO.Create(g); err != nil {
		return nil, err
	}
	_ = s.groupDAO.AddMember(g.ID, ownerID)
	for _, uid := range memberIDs {
		if uid != ownerID {
			_ = s.groupDAO.AddMember(g.ID, uid)
		}
	}
	return s.groupDAO.GetByID(g.ID)
}

// ListMyGroups 我加入的群列表
func (s *GroupService) ListMyGroups(userID int64) ([]GroupDTO, error) {
	groups, err := s.groupDAO.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make([]GroupDTO, 0, len(groups))
	for _, g := range groups {
		cnt, _ := s.groupDAO.MemberCount(g.ID)
		last, _ := s.msgDAO.LastGroupMessage(g.ID)
		result = append(result, GroupDTO{Group: g, MemberCount: cnt, LastMessage: last})
	}
	return result, nil
}

// GetGroup 获取群详情（需是成员）
func (s *GroupService) GetGroup(userID, groupID int64) (*model.Group, []model.GroupMember, error) {
	ok, err := s.groupDAO.IsMember(groupID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, ErrNotGroupMember
	}
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return nil, nil, err
	}
	members, err := s.groupDAO.GetMembers(groupID)
	if err != nil {
		return nil, nil, err
	}
	return g, members, nil
}

// SearchGroups 搜索群组
func (s *GroupService) SearchGroups(keyword string) ([]model.Group, error) {
	return s.groupDAO.Search(keyword)
}

// JoinGroup 加入群组
func (s *GroupService) JoinGroup(userID, groupID int64) error {
	ok, err := s.groupDAO.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if ok {
		return ErrAlreadyMember
	}
	if _, err := s.groupDAO.GetByID(groupID); err != nil {
		return errors.New("群组不存在")
	}
	return s.groupDAO.AddMember(groupID, userID)
}

// LeaveGroup 退出群组
func (s *GroupService) LeaveGroup(userID, groupID int64) error {
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return errors.New("群组不存在")
	}
	if g.OwnerID == userID {
		return errors.New("群主不能退出群组，请先解散群组")
	}
	return s.groupDAO.RemoveMember(groupID, userID)
}

// KickMember 群主踢人
func (s *GroupService) KickMember(ownerID, groupID, targetID int64) error {
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return errors.New("群组不存在")
	}
	if g.OwnerID != ownerID {
		return ErrNotGroupOwner
	}
	if targetID == ownerID {
		return errors.New("不能踢出自己")
	}
	return s.groupDAO.RemoveMember(groupID, targetID)
}

// DisbandGroup 解散群组
func (s *GroupService) DisbandGroup(userID, groupID int64) error {
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return errors.New("群组不存在")
	}
	if g.OwnerID != userID {
		return ErrNotGroupOwner
	}
	return s.groupDAO.Delete(groupID)
}

// SendGroupMessage 发送群消息
func (s *GroupService) SendGroupMessage(fromID, groupID int64, msgType string, content interface{}) (*model.Message, error) {
	ok, err := s.groupDAO.IsMember(groupID, fromID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotGroupMember
	}
	contentStr, err := marshalContent(content)
	if err != nil {
		return nil, err
	}
	from, _ := s.userDAO.GetByID(fromID)
	msg := &model.Message{
		FromID:    fromID,
		ToID:      groupID,
		ChatType:  "group",
		MsgType:   msgType,
		Content:   contentStr,
		IsRead:    true,
		CreatedAt: time.Now(),
		From:      from,
	}
	if err := s.msgDAO.Save(msg); err != nil {
		return nil, err
	}
	// 推送给所有在线群成员（除发送者外）
	members, _ := s.groupDAO.GetMembers(groupID)
	for _, m := range members {
		if m.UserID != fromID {
			s.hub.SendToUser(m.UserID, "message", msg)
		}
	}
	return msg, nil
}

// GetGroupHistory 获取群消息历史
func (s *GroupService) GetGroupHistory(userID, groupID int64, beforeID int64, limit int) ([]model.Message, error) {
	ok, err := s.groupDAO.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotGroupMember
	}
	if limit <= 0 || limit > 50 {
		limit = 30
	}
	return s.msgDAO.GetGroupHistory(groupID, beforeID, limit)
}

type GroupDTO struct {
	model.Group
	MemberCount int64          `json:"member_count"`
	LastMessage *model.Message `json:"last_message"`
}

// UpdateGroupAvatar 群主更换群头像，返回旧头像路径（供调用方删除文件）
func (s *GroupService) UpdateGroupAvatar(ownerID, groupID int64, avatarURL string) (string, error) {
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return "", errors.New("群组不存在")
	}
	if g.OwnerID != ownerID {
		return "", ErrNotGroupOwner
	}
	oldAvatar := g.Avatar
	if err := s.groupDAO.UpdateAvatar(groupID, avatarURL); err != nil {
		return "", err
	}
	return oldAvatar, nil
}

// UpdateSettings 群主修改群设置
func (s *GroupService) UpdateSettings(userID, groupID int64, allowInvite bool) error {
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return errors.New("群组不存在")
	}
	if g.OwnerID != userID {
		return ErrNotGroupOwner
	}
	return s.groupDAO.UpdateSettings(groupID, allowInvite)
}

// InviteToGroup 邀请用户加群
func (s *GroupService) InviteToGroup(inviterID, groupID, inviteeID int64) (*model.GroupInvite, error) {
	g, err := s.groupDAO.GetByID(groupID)
	if err != nil {
		return nil, errors.New("群组不存在")
	}
	// 权限检查：群主 or (allow_invite=true 且邀请人是成员)
	ok, _ := s.groupDAO.IsMember(groupID, inviterID)
	if !ok {
		return nil, ErrNotGroupMember
	}
	if g.OwnerID != inviterID && !g.AllowInvite {
		return nil, errors.New("该群不允许成员邀请")
	}
	// 被邀请者已是成员
	alreadyIn, _ := s.groupDAO.IsMember(groupID, inviteeID)
	if alreadyIn {
		return nil, ErrAlreadyMember
	}
	// 已有 pending 邀请
	if existing, _ := s.groupDAO.GetPendingInvite(groupID, inviteeID); existing != nil {
		return nil, errors.New("已发送过邀请，等待对方确认")
	}
	inv := &model.GroupInvite{
		GroupID:   groupID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    "pending",
	}
	if err := s.groupDAO.CreateInvite(inv); err != nil {
		return nil, err
	}
	// 加载关联数据后推送 WS
	full, _ := s.groupDAO.GetInvite(inv.ID)
	if full != nil {
		s.hub.SendToUser(inviteeID, "group_invite", full)
	}
	return inv, nil
}

// HandleInvite 接受或拒绝邀请
func (s *GroupService) HandleInvite(userID, inviteID int64, accept bool) error {
	inv, err := s.groupDAO.GetInvite(inviteID)
	if err != nil {
		return ErrInviteNotFound
	}
	if inv.InviteeID != userID {
		return ErrInviteNotForYou
	}
	if inv.Status != "pending" {
		return ErrInviteNotPending
	}
	if accept {
		if err := s.groupDAO.UpdateInviteStatus(inviteID, "accepted"); err != nil {
			return err
		}
		return s.groupDAO.AddMember(inv.GroupID, userID)
	}
	return s.groupDAO.UpdateInviteStatus(inviteID, "rejected")
}

// ListMyInvites 获取我的待处理邀请
func (s *GroupService) ListMyInvites(userID int64) ([]model.GroupInvite, error) {
	return s.groupDAO.ListPendingInvitesForUser(userID)
}

// RecallGroupMessage 撤回群聊消息（本人或群主均可）
func (s *GroupService) RecallGroupMessage(userID, msgID int64) error {
	msg, err := s.msgDAO.GetByID(msgID)
	if err != nil || msg == nil {
		return ErrMsgNotFound
	}
	if msg.ChatType != "group" {
		return ErrRecallForbidden
	}
	if time.Since(msg.CreatedAt) > recallWindow {
		return ErrRecallTimeout
	}
	// 只有消息发送者或群主可以撤回
	if msg.FromID != userID {
		g, err := s.groupDAO.GetByID(msg.ToID)
		if err != nil || g == nil || g.OwnerID != userID {
			return ErrRecallForbidden
		}
	}
	if err := s.msgDAO.Recall(msgID); err != nil {
		return err
	}
	// 广播给群内所有在线成员
	members, _ := s.groupDAO.GetMembers(msg.ToID)
	payload := map[string]any{"msg_id": msgID, "chat_type": "group", "to_id": msg.ToID, "from_id": msg.FromID}
	for _, m := range members {
		s.hub.SendToUser(m.UserID, "message_recalled", payload)
	}
	return nil
}
