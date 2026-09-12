package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/middleware"
	"github.com/lp/campus-market/internal/service"
	"github.com/lp/campus-market/internal/util"
)

// ConversationHandler exposes private-message endpoints.
type ConversationHandler struct {
	svc     *service.ConversationService
	products *service.ProductService
	users   *service.UserService
	logger  *slog.Logger
}

// NewConversationHandler wires the conversation handler dependencies.
func NewConversationHandler(svc *service.ConversationService, products *service.ProductService, users *service.UserService, logger *slog.Logger) *ConversationHandler {
	return &ConversationHandler{svc: svc, products: products, users: users, logger: logger}
}

// Create handles POST /conversations.
func (h *ConversationHandler) Create(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	user, err := h.users.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	product, err := h.products.Get(c.Request.Context(), req.ProductID)
	if err != nil {
		c.Error(err)
		return
	}
	conv, err := h.svc.Create(c.Request.Context(), user, product.SellerID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, conv)
}

// ListMy handles GET /conversations/me.
func (h *ConversationHandler) ListMy(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	items, err := h.svc.ListMy(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, items)
}

// SendMessage handles POST /conversations/:id/messages.
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "会话ID不合法")
		return
	}
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	msg, err := h.svc.SendMessage(c.Request.Context(), userID, uint(convID), req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, msg)
}

// ListMessages handles GET /conversations/:id/messages.
func (h *ConversationHandler) ListMessages(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "会话ID不合法")
		return
	}
	items, err := h.svc.ListMessages(c.Request.Context(), userID, uint(convID))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, items)
}
