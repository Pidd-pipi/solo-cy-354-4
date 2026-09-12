package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/handler"
)

// RegisterUserRoutes registers student endpoints.
func RegisterUserRoutes(g *gin.RouterGroup, h *handler.UserHandler, auth gin.HandlerFunc, loginLimiter, apiLimiter gin.HandlerFunc) {
	users := g.Group("/users")
	{
		users.POST("/register", loginLimiter, h.Register)
		users.POST("/login", loginLimiter, h.Login)
		me := users.Group("/me", auth)
		{
			me.GET("", apiLimiter, h.GetProfile)
			me.PUT("", apiLimiter, h.UpdateProfile)
		}
	}
}
