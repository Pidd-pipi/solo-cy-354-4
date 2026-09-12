package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/handler"
)

// RegisterProductRoutes registers product endpoints.
func RegisterProductRoutes(g *gin.RouterGroup, h *handler.ProductHandler, auth, apiLimiter gin.HandlerFunc) {
	products := g.Group("/products")
	{
		products.GET("", apiLimiter, h.List)
		products.GET("/graduation", apiLimiter, h.Graduation)
		products.GET("/:id", apiLimiter, h.Get)
		authed := products.Group("", auth)
		{
			authed.POST("", apiLimiter, h.Create)
			authed.DELETE("/:id", apiLimiter, h.Remove)
		}
	}
}
