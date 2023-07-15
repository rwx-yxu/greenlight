package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
)

func RegisterUserHandler(c *gin.Context, app app.Application) {
	// Create an anonymous struct to hold the expected data from the request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body into the anonymous struct.
	err := ReadJSON(c, &input)
	if err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}

	// Copy the data from the request body into a new User struct. Notice also that we
	// set the Activated field to false, which isn't strictly necessary because the
	// Activated field will have the zero-value of false by default. But setting this
	// explicitly helps to make our intentions clear to anyone reading the code.
	user := &models.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Use the Password.Set() method to generate and store the hashed and plaintext
	// passwords.
	err = user.Password.Set(input.Password)
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}
	v, err := app.User.Add(user)
	if !v.Valid() {
		ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		return
	}
	if err != nil {
		switch {
		case errors.Is(err, brokers.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		default:
			ErrorResponse(c, app, InternalServerError(err))
		}
		return
	}
	err = app.SMTP.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}
