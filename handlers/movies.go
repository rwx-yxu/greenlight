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
		Title   *string         `json:"title"`
		Year    *int32          `json:"year"`
		Runtime *models.Runtime `json:"runtime"`
		Genres  []string        `json:"genres"`
	}
	if err := ReadJSON(c, &input); err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}
	// If the input.Title value is nil then we know that no corresponding "title" key/
	// value pair was provided in the JSON request body. So we move on and leave the
	// movie record unchanged. Otherwise, we update the movie record with the new title
	// value. Importantly, because input.Title is a now a pointer to a string, we need
	// to dereference the pointer using the * operator to get the underlying value
	// before assigning it to our movie record.
	if input.Title != nil {
		movie.Title = *input.Title
	}

	// We also do the same for the other fields in the input struct.
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres // Note that we don't need to dereference a slice.
	}

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

func DeleteMovieHandler(c *gin.Context, app app.Application) {
	// Extract the movie ID from the URL.
	id, err := ReadIDParam(c)
	if err != nil {
		ErrorResponse(c, app, NotFoundError(err))
		return
	}

	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.Movie.RemoveByID(id)
	if err != nil {
		switch {
		case errors.Is(err, brokers.ErrRecordNotFound):
			ErrorResponse(c, app, NotFoundError(err))
		default:
			ErrorResponse(c, app, InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "movie successfully deleted"})
}
