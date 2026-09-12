package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/handler"
)

// RegisterReviewRoutes registers review endpoints.
func RegisterReviewRoutes(g *gin.RouterGroup, h *handler.ReviewHandler, auth, apiLimiter gin.HandlerFunc) {
	reviews := g.Group("/reviews", auth)
	{
		reviews.POST("", apiLimiter, h.Create)
		reviews.GET("/me", apiLimiter, h.ListMine)
	}
}
