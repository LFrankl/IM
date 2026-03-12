package router

import (
	"im-backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func stub(c *gin.Context) {
	response.Fail(c, http.StatusNotImplemented, "not implemented yet")
}
