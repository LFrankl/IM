package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Fail(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, "unauthorized")
}

func Forbidden(c *gin.Context) {
	Fail(c, http.StatusForbidden, "forbidden")
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, message)
}

func InternalError(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, "internal server error")
}
