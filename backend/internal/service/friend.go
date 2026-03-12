package service

import (
	"errors"
	"im-backend/internal/dao"
	"im-backend/internal/model"
	"im-backend/internal/ws"
)

type FriendService struct {
	friendDAO *dao.FriendDAO
	userDAO   *dao.UserDAO
	hub       *ws.Hub
}

func NewFriendService(friendDAO *dao.FriendDAO, userDAO *dao.UserDAO, hub *ws.Hub) *FriendService {
	return &FriendService{friendDAO: friendDAO, userDAO: userDAO, hub: hub}
}

var (
	ErrAlreadyFriend      = errors.New("已经是好友了")
	ErrRequestSelf        = errors.New("不能添加自己为好友")
	ErrRequestNotFound    = errors.New("申请不存在")
	ErrRequestNotPending  = errors.New("申请已处理")
	ErrRequestNotYours    = errors.New("无权操作此申请")
	ErrNotFriend          = errors.New("对方不是你的好友")
)

// ListFriends 获取好友列表
func (s *FriendService) ListFriends(userID int64) ([]model.Friendship, error) {
	return s.friendDAO.ListFriends(userID)
}

// SendRequest 发送好友申请
func (s *FriendService) SendRequest(fromID, toID int64, message string) (*model.FriendRequest, error) {
	if fromID == toID {
		return nil, ErrRequestSelf
	}
	// 目标用户存在？
	target, err := s.userDAO.GetByID(toID)
	if err != nil || target == nil {
		return nil, ErrUserNotFound
	}
	// 已经是好友？
	ok, err := s.friendDAO.IsFriend(fromID, toID)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ErrAlreadyFriend
	}

	req, err := s.friendDAO.CreateRequest(fromID, toID, message)
	if err != nil {
		return nil, err
	}

	// 如果对方在线，推送通知
	from, _ := s.userDAO.GetByID(fromID)
	s.hub.SendToUser(toID, "friend_request", map[string]any{
		"request": req,
		"from":    from,
	})

	return req, nil
}

// ListPendingRequests 收到的待处理申请
func (s *FriendService) ListPendingRequests(userID int64) ([]model.FriendRequest, error) {
	return s.friendDAO.ListPendingReceived(userID)
}

// HandleRequest 同意或拒绝申请
func (s *FriendService) HandleRequest(userID, reqID int64, accept bool) error {
	req, err := s.friendDAO.GetRequestByID(reqID)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrRequestNotFound
	}
	if req.ToID != userID {
		return ErrRequestNotYours
	}
	if req.Status != "pending" {
		return ErrRequestNotPending
	}

	if accept {
		if err := s.friendDAO.AddFriendship(req.FromID, req.ToID); err != nil {
			return err
		}
		if err := s.friendDAO.UpdateRequestStatus(reqID, "accepted"); err != nil {
			return err
		}
		// 通知申请方已被接受
		toUser, _ := s.userDAO.GetByID(req.ToID)
		s.hub.SendToUser(req.FromID, "friend_accepted", map[string]any{
			"request_id": reqID,
			"user":       toUser,
		})
	} else {
		if err := s.friendDAO.UpdateRequestStatus(reqID, "rejected"); err != nil {
			return err
		}
	}
	return nil
}

// DeleteFriend 删除好友
func (s *FriendService) DeleteFriend(userID, friendID int64) error {
	ok, err := s.friendDAO.IsFriend(userID, friendID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFriend
	}
	return s.friendDAO.DeleteFriendship(userID, friendID)
}

// UpdateRemark 修改备注
func (s *FriendService) UpdateRemark(userID, friendID int64, remark string) error {
	ok, err := s.friendDAO.IsFriend(userID, friendID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFriend
	}
	return s.friendDAO.UpdateRemark(userID, friendID, remark)
}

// UpdateGroup 修改分组
func (s *FriendService) UpdateGroup(userID, friendID int64, groupName string) error {
	ok, err := s.friendDAO.IsFriend(userID, friendID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFriend
	}
	return s.friendDAO.UpdateGroupName(userID, friendID, groupName)
}

// CountPendingRequests 待处理申请数（角标用）
func (s *FriendService) CountPendingRequests(userID int64) (int64, error) {
	return s.friendDAO.CountPendingReceived(userID)
}
