package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/handler"
)

// RegisterBookExchangeRoutes registers book swap endpoints.
func RegisterBookExchangeRoutes(g *gin.RouterGroup, h *handler.BookExchangeHandler, auth, apiLimiter gin.HandlerFunc) {
	exchanges := g.Group("/book-exchanges")
	{
		exchanges.GET("", apiLimiter, h.List)
		authed := exchanges.Group("", auth)
		{
			authed.POST("", apiLimiter, h.Create)
			authed.POST("/:id/close", apiLimiter, h.Close)
		}
	}
}
