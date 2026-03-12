package main

import (
	"fmt"
	"im-backend/internal/config"
	"im-backend/internal/dao"
	"im-backend/internal/handler"
	"im-backend/internal/model"
	"im-backend/internal/router"
	"im-backend/internal/service"
	"im-backend/internal/ws"
	pkgdb "im-backend/pkg/database"
	pkgjwt "im-backend/pkg/jwt"
	"log"
	"os"
	"time"
)

func main() {
	cfgPath := "config/config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		cfgPath = p
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pkgjwt.Init(cfg.JWT.Secret)

	db, err := pkgdb.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Friendship{},
		&model.FriendRequest{},
		&model.Message{},
		&model.Group{},
		&model.GroupMember{},
		&model.SpacePost{},
		&model.SpaceComment{},
		&model.SpaceLike{},
	); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	expire, err := time.ParseDuration(cfg.JWT.Expire)
	if err != nil {
		log.Fatalf("parse jwt expire: %v", err)
	}

	// DAO
	userDAO := dao.NewUserDAO(db)
	friendDAO := dao.NewFriendDAO(db)
	msgDAO := dao.NewMessageDAO(db)
	groupDAO := dao.NewGroupDAO(db)
	spaceDAO := dao.NewSpaceDAO(db)

	// Service
	authSvc := service.NewAuthService(userDAO, expire)
	userSvc := service.NewUserService(userDAO)
	friendSvc := service.NewFriendService(friendDAO, userDAO, ws.Global)
	chatSvc := service.NewChatService(msgDAO, friendDAO, userDAO, ws.Global)
	groupSvc := service.NewGroupService(groupDAO, msgDAO, userDAO, ws.Global)
	spaceSvc := service.NewSpaceService(spaceDAO, friendDAO, userDAO)

	// Handler
	handlers := &router.Handlers{
		Auth:   handler.NewAuthHandler(authSvc, userSvc),
		User:   handler.NewUserHandler(userSvc),
		Friend: handler.NewFriendHandler(friendSvc),
		Chat:   handler.NewChatHandler(chatSvc),
		Group:  handler.NewGroupHandler(groupSvc),
		Space:  handler.NewSpaceHandler(spaceSvc),
		WS:     handler.NewWSHandler(chatSvc, groupSvc, ws.Global),
	}

	r := router.New(cfg.CORS.AllowOrigins, handlers)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("IM server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
