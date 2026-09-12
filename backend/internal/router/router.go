// Package router assembles the Gin engine and all route groups for campus-market.
package router

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/config"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/handler"
	"github.com/lp/campus-market/internal/middleware"
	"github.com/lp/campus-market/internal/repository"
	"github.com/lp/campus-market/internal/service"
	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

// New builds the Gin engine with all dependencies wired.
func New(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID(logger))
	r.Use(middleware.AccessLog(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:  cfg.CORSOrigins,
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"X-Request-ID"},
	}))
	r.Use(middleware.ErrorHandler())

	r.GET("/healthz", func(c *gin.Context) { util.OK(c, gin.H{"status": "ok"}) })

	// repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	convRepo := repository.NewConversationRepository(db)
	orderRepo := repository.NewTradeOrderRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	exchangeRepo := repository.NewBookExchangeRepository(db)

	// services
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours, logger)
	productSvc := service.NewProductService(productRepo, logger)
	convSvc := service.NewConversationService(convRepo, logger)
	orderSvc := service.NewTradeOrderService(orderRepo, productRepo, logger)
	reviewSvc := service.NewReviewService(reviewRepo, orderRepo, userRepo, logger)
	exchangeSvc := service.NewBookExchangeService(exchangeRepo, logger)

	// handlers
	userH := handler.NewUserHandler(userSvc, logger)
	productH := handler.NewProductHandler(productSvc, logger)
	convH := handler.NewConversationHandler(convSvc, productSvc, userSvc, logger)
	orderH := handler.NewTradeOrderHandler(orderSvc, userSvc, logger)
	reviewH := handler.NewReviewHandler(reviewSvc, userSvc, logger)
	exchangeH := handler.NewBookExchangeHandler(exchangeSvc, logger)

	auth := middleware.AuthRequired(cfg.JWTSecret, logger)
	requireAdmin := middleware.RequireRole(logger, constants.UserRoleAdmin)
	loginLimiter := middleware.RateLimit(middleware.NewRateLimiter(cfg.LoginRateLimit, time.Minute), logger)
	apiLimiter := middleware.RateLimit(middleware.NewRateLimiter(cfg.RateLimitPerMin, time.Minute), logger)

	v1 := r.Group("/api/v1")
	{
		RegisterUserRoutes(v1, userH, auth, loginLimiter, apiLimiter)
		RegisterProductRoutes(v1, productH, auth, apiLimiter)
		RegisterConversationRoutes(v1, convH, auth, apiLimiter)
		RegisterTradeOrderRoutes(v1, orderH, auth, apiLimiter)
		RegisterReviewRoutes(v1, reviewH, auth, apiLimiter)
		RegisterBookExchangeRoutes(v1, exchangeH, auth, apiLimiter)
		// admin-only report handling placeholder route group (kept for RBAC coverage)
		admin := v1.Group("/admin", auth, requireAdmin)
		admin.GET("/stats", func(c *gin.Context) { util.OK(c, gin.H{"users": "admin-stats"}) })
	}
	return r
}
