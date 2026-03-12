package middleware

import (
	"im-backend/pkg/jwt"
	"im-backend/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const UsernameKey = "username"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先读 Authorization header，WS 握手时退回 query string
		tokenStr := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		claims, err := jwt.Parse(tokenStr)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	id, _ := c.Get(UserIDKey)
	userID, _ := id.(int64)
	return userID
}
