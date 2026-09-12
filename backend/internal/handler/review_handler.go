package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/middleware"
	"github.com/lp/campus-market/internal/service"
	"github.com/lp/campus-market/internal/util"
)

// ReviewHandler exposes review endpoints.
type ReviewHandler struct {
	svc    *service.ReviewService
	users  *service.UserService
	logger *slog.Logger
}

// NewReviewHandler wires the review handler dependencies.
func NewReviewHandler(svc *service.ReviewService, users *service.UserService, logger *slog.Logger) *ReviewHandler {
	return &ReviewHandler{svc: svc, users: users, logger: logger}
}

// Create handles POST /reviews.
func (h *ReviewHandler) Create(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	user, err := h.users.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	rv, err := h.svc.Create(c.Request.Context(), user, &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, rv)
}

// ListMine handles GET /reviews/me.
func (h *ReviewHandler) ListMine(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	items, err := h.svc.ListByReviewee(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, items)
}
