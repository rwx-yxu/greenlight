package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rwx-yxu/greenlight/app"
)

func CreateMovieHandler(c *gin.Context, app app.Application) {
	c.String(http.StatusOK, "create a new movie")
}

func ShowMovieHandler(c *gin.Context, app app.Application) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("movie uuid:%s\n", id))
}
