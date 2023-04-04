package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/internal/services/movie"
)

func CreateMovieHandler(c *gin.Context, app app.Application) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}
	maxBytes := int64(1048576)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
	if err := ReadJSON(c, &input); err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"movie": input})

}

func ShowMovieHandler(c *gin.Context, app app.Application) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, app, NotFoundError(err))
		return
	}
	movie := movie.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}
