package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/handler"
)

// RegisterTradeOrderRoutes registers trade order endpoints.
func RegisterTradeOrderRoutes(g *gin.RouterGroup, h *handler.TradeOrderHandler, auth, apiLimiter gin.HandlerFunc) {
	orders := g.Group("/trade-orders", auth)
	{
		orders.POST("", apiLimiter, h.Create)
		orders.GET("/me", apiLimiter, h.ListMy)
		orders.POST("/:id/buyer-confirm", apiLimiter, h.BuyerConfirm)
		orders.POST("/:id/seller-confirm", apiLimiter, h.SellerConfirm)
		orders.POST("/:id/cancel", apiLimiter, h.Cancel)
	}
}
