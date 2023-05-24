package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
)

func CreateMovieHandler(c *gin.Context, app app.Application) {
	var input struct {
		Title   string         `json:"title"`
		Year    int32          `json:"year"`
		Runtime models.Runtime `json:"runtime"`
		Genres  []string       `json:"genres"`
	}
	if err := ReadJSON(c, &input); err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}

	m := &models.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v, err := app.Movie.Add(m)
	if v != nil {
		ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		return
	}
	if err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}
	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	c.Header("Location", fmt.Sprintf("/v1/movies/%d", m.ID))
	c.JSON(http.StatusCreated, gin.H{"movie": m})

}

func ShowMovieHandler(c *gin.Context, app app.Application) {
	id, err := ReadIDParam(c)
	if err != nil {
		ErrorResponse(c, app, NotFoundError(err))
		return
	}

	movie, err := app.Movie.FindByID(id)
	if err != nil {
		switch {
		case errors.Is(err, brokers.ErrRecordNotFound):
			ErrorResponse(c, app, NotFoundError(err))
		default:
			ErrorResponse(c, app, InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

func UpdateMovieHandler(c *gin.Context, app app.Application) {
	id, err := ReadIDParam(c)
	if err != nil {
		ErrorResponse(c, app, NotFoundError(err))
		return
	}
	movie, err := app.Movie.FindByID(id)
	if err != nil {
		switch {
		case errors.Is(err, brokers.ErrRecordNotFound):
			ErrorResponse(c, app, NotFoundError(err))
		default:
			ErrorResponse(c, app, InternalServerError(err))
		}
		return
	}
	var input struct {
		Title   string         `json:"title"`
		Year    int32          `json:"year"`
		Runtime models.Runtime `json:"runtime"`
		Genres  []string       `json:"genres"`
	}
	if err := ReadJSON(c, &input); err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}
	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	v, err := app.Movie.Edit(movie)
	if v != nil {
		ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		return
	}
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"movie": movie})
}
