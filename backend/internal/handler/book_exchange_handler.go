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

// BookExchangeHandler exposes book swap endpoints.
type BookExchangeHandler struct {
	svc    *service.BookExchangeService
	logger *slog.Logger
}

// NewBookExchangeHandler wires the book exchange handler dependencies.
func NewBookExchangeHandler(svc *service.BookExchangeService, logger *slog.Logger) *BookExchangeHandler {
	return &BookExchangeHandler{svc: svc, logger: logger}
}

// Create handles POST /book-exchanges.
func (h *BookExchangeHandler) Create(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	var req dto.CreateBookExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	e, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, e)
}

// List handles GET /book-exchanges.
func (h *BookExchangeHandler) List(c *gin.Context) {
	var q dto.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	result, err := h.svc.List(c.Request.Context(), &q)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Close handles POST /book-exchanges/:id/close.
func (h *BookExchangeHandler) Close(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "交换ID不合法")
		return
	}
	e, err := h.svc.Close(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, e)
}
