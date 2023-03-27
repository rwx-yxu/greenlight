package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/handlers"
)

func MethodNotAllowed(app app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Writer.Status() == http.StatusMethodNotAllowed {
			handlers.ErrorResponse(c, app, handlers.NotAllowedError(nil, c.Request.Method))
			return
		}

		c.Next()
	}
}

func NotFound(app app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Writer.Status() == http.StatusNotFound {
			handlers.ErrorResponse(c, app, handlers.NotFoundError(nil))
			return
		}

		c.Next()
	}
}
