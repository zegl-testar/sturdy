package routes

import (
	"getsturdy.com/api/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthDestroy(c *gin.Context) {
	auth.RemoveAuthCookie(c.Writer)
	c.Status(http.StatusOK)
}
