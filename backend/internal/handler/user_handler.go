// Package handler implements the HTTP handlers for campus-market.
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

// UserHandler exposes student endpoints.
type UserHandler struct {
	svc    *service.UserService
	logger *slog.Logger
}

// NewUserHandler wires the user handler dependencies.
func NewUserHandler(svc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// Register handles POST /users/register.
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	user, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, service.ToUserView(user))
}

// Login handles POST /users/login.
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, resp)
}

// GetProfile handles GET /users/me.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, service.ToUserView(user))
}

// UpdateProfile handles PUT /users/me.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, service.ToUserView(user))
}
