package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
)

func HealthcheckHandler(c *gin.Context, app app.Application) {
	c.JSON(200, gin.H{
		"status":      "available",
		"environment": app.Config.Env,
		"version":     app.Config.Version,
	})
}
