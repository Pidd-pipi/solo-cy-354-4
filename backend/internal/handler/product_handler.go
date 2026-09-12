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

// ProductHandler exposes second-hand product endpoints.
type ProductHandler struct {
	svc    *service.ProductService
	logger *slog.Logger
}

// NewProductHandler wires the product handler dependencies.
func NewProductHandler(svc *service.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

// Create handles POST /products.
func (h *ProductHandler) Create(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, constants.MsgValidationFailed)
		return
	}
	p, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, p)
}

// Get handles GET /products/:id.
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "商品ID不合法")
		return
	}
	p, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, p)
}

// List handles GET /products.
func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ListProductQuery
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

// Remove handles DELETE /products/:id.
func (h *ProductHandler) Remove(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "商品ID不合法")
		return
	}
	p, err := h.svc.Remove(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, p)
}

// Graduation handles GET /products/graduation (毕业季专场).
func (h *ProductHandler) Graduation(c *gin.Context) {
	result, err := h.svc.List(c.Request.Context(), &dto.ListProductQuery{PageQuery: dto.PageQuery{Page: 1, PageSize: 50}, Status: constants.ProductStatusOnSale})
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}
