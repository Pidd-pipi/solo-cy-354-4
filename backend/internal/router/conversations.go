package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/handler"
)

// RegisterConversationRoutes registers private-message endpoints.
func RegisterConversationRoutes(g *gin.RouterGroup, h *handler.ConversationHandler, auth, apiLimiter gin.HandlerFunc) {
	convs := g.Group("/conversations", auth)
	{
		convs.POST("", apiLimiter, h.Create)
		convs.GET("/me", apiLimiter, h.ListMy)
		convs.GET("/:id/messages", apiLimiter, h.ListMessages)
		convs.POST("/:id/messages", apiLimiter, h.SendMessage)
	}
}
