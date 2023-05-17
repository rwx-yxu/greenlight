package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
)

func HealthcheckHandler(c *gin.Context, app app.Application) {
	c.JSON(200, gin.H{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.Config.Server.Env,
			"version":     app.Config.Server.Version,
		},
	})
}
