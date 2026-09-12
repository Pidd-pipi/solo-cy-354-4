package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/util"
)

// ErrorHandler converts panics and handler errors into the unified envelope.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		switch {
		case errors.As(err, &appErr):
			util.Fail(c, appErr.Status, appErr.Code, appErr.Message)
		case errors.Is(err, util.ErrNotFound):
			util.Fail(c, http.StatusNotFound, constants.CodeNotFound, constants.MsgNotFound)
		case errors.Is(err, util.ErrConflict):
			util.Fail(c, http.StatusConflict, constants.CodeConflict, constants.MsgConflict)
		default:
			util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
		}
	}
}
