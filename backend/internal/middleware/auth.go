package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/util"
)

// AuthRequired validates the JWT and stores the user id/role into the context.
func AuthRequired(secret string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token == "" {
			logger.Warn(fmt.Sprintf(constants.LogMiddlewareAuthFailed, "missing token"))
			util.Fail(c, 401, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		claims, err := util.ParseToken(secret, token)
		if err != nil {
			logger.Warn(fmt.Sprintf(constants.LogMiddlewareAuthFailed, err))
			util.Fail(c, 401, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_phone", claims.Phone)
		c.Next()
	}
}

// CurrentUserID reads the authenticated user id from the context.
func CurrentUserID(c *gin.Context) (uint, error) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, errors.New("user id missing")
	}
	switch id := v.(type) {
	case uint:
		return id, nil
	case float64:
		return uint(id), nil
	case string:
		n, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(n), nil
	default:
		return 0, errors.New("user id invalid")
	}
}

// CurrentRole reads the authenticated user role from the context.
func CurrentRole(c *gin.Context) string {
	v, ok := c.Get("user_role")
	if !ok {
		return ""
	}
	role, _ := v.(string)
	return role
}
