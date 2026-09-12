package middleware

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/util"
)

// RequireRole rejects requests whose role is not in the allowed set.
func RequireRole(logger *slog.Logger, allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := CurrentRole(c)
		for _, a := range allowed {
			if role == a {
				c.Next()
				return
			}
		}
		userID, _ := CurrentUserID(c)
		logger.Warn(fmt.Sprintf(constants.LogMiddlewareRbacDenied, userID, role, allowed))
		util.Fail(c, 403, constants.CodeForbidden, constants.MsgForbidden)
		c.Abort()
	}
}
