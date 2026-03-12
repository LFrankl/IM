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
	ErrNotGroupMember = errors.New("你不是该群成员")
	ErrNotGroupOwner  = errors.New("只有群主才能执行此操作")
	ErrAlreadyMember  = errors.New("已经是群成员")
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
