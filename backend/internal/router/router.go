package router

import (
	"im-backend/internal/handler"
	"im-backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Friend *handler.FriendHandler
	Chat   *handler.ChatHandler
	Group  *handler.GroupHandler
	Space  *handler.SpaceHandler
	WS     *handler.WSHandler
}

func New(allowOrigins []string, h *Handlers) *gin.Engine {
	r := gin.Default()

	// 本地单机部署：允许所有 localhost/127.0.0.1 来源（防止 Vite 端口变化导致 403）
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(cfg))

	r.Static("/uploads", "./data/uploads")

	api := r.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
	}

	authed := api.Group("", middleware.Auth())
	{
		authed.POST("/auth/logout", h.Auth.Logout)
		authed.GET("/auth/me", h.Auth.Me)

		authed.GET("/users/search", h.User.Search)
		authed.GET("/users/:id", h.User.GetUser)
		authed.PUT("/users/me", h.User.UpdateProfile)
		authed.POST("/users/me/avatar", h.User.UploadAvatar)
		authed.POST("/users/me/cover", h.User.UploadCover)

		authed.GET("/friends", h.Friend.ListFriends)
		authed.POST("/friends/requests", h.Friend.SendRequest)
		authed.GET("/friends/requests", h.Friend.ListRequests)
		authed.GET("/friends/requests/count", h.Friend.CountPending)
		authed.PUT("/friends/requests/:id", h.Friend.HandleRequest)
		authed.DELETE("/friends/:id", h.Friend.DeleteFriend)
		authed.PUT("/friends/:id/remark", h.Friend.UpdateRemark)
		authed.PUT("/friends/:id/group", h.Friend.UpdateGroup)

		authed.GET("/conversations", h.Chat.ListConversations)
		authed.GET("/messages/:userId", h.Chat.GetMessages)
		authed.PUT("/messages/:userId/read", h.Chat.MarkRead)
		authed.POST("/messages/upload", h.Chat.UploadFile)

		authed.GET("/groups", h.Group.ListMyGroups)
		authed.POST("/groups", h.Group.CreateGroup)
		authed.GET("/groups/search", h.Group.SearchGroups)
		authed.GET("/groups/invites", h.Group.ListMyInvites)
		authed.PUT("/groups/invites/:inviteId", h.Group.HandleInvite)
		authed.GET("/groups/:id", h.Group.GetGroup)
		authed.GET("/groups/:id/members", h.Group.GetMembers)
		authed.POST("/groups/:id/join", h.Group.JoinGroup)
		authed.DELETE("/groups/:id/leave", h.Group.LeaveGroup)
		authed.DELETE("/groups/:id/members/:uid", h.Group.KickMember)
		authed.DELETE("/groups/:id", h.Group.DisbandGroup)
		authed.GET("/groups/:id/messages", h.Group.GetMessages)
		authed.PUT("/groups/:id/settings", h.Group.UpdateSettings)
		authed.POST("/groups/:id/avatar", h.Group.UpdateAvatar)
		authed.POST("/groups/:id/invites", h.Group.InviteMember)

		authed.GET("/space/feed", h.Space.GetFeed)
		authed.GET("/space/users/:id/posts", h.Space.GetUserPosts)
		authed.POST("/space/posts", h.Space.CreatePost)
		authed.DELETE("/space/posts/:id", h.Space.DeletePost)
		authed.POST("/space/posts/:id/like", h.Space.LikePost)
		authed.DELETE("/space/posts/:id/like", h.Space.UnlikePost)
		authed.POST("/space/posts/:id/comments", h.Space.AddComment)
		authed.DELETE("/space/comments/:id", h.Space.DeleteComment)
	}

	r.GET("/ws", middleware.Auth(), h.WS.Handle)

	return r
}
