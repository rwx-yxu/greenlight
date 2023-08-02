package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
	"github.com/rwx-yxu/greenlight/internal/validator"
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

	err = app.Permission.AddForUser(user.ID, "movies:read")
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}

	token, err := models.GenerateToken(user.ID, 3*24*time.Hour, models.ScopeActivation)
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}

	v, err = app.Token.Add(token)
	if !v.Valid() {
		ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		return
	}
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}
	app.Background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		err = app.SMTP.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.Logger.PrintError(err, nil)
		}
	})
	c.JSON(http.StatusAccepted, gin.H{"user": user})
}

func ActivateUserHandler(c *gin.Context, app app.Application) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	if err := ReadJSON(c, &input); err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}

	v := validator.New()
	if app.Token.ValidatePlainText(v, input.TokenPlaintext); !v.Valid() {
		ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		return
	}
	user, err := app.User.FindByToken(models.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, brokers.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			ErrorResponse(c, app, FailedValidationResponse(v.Errors))
		default:
			ErrorResponse(c, app, InternalServerError(err))
		}
		return
	}
	user.Activated = true
	//Update user. Do not need validator return value because the user model has already been
	//validated when finding the token
	_, err = app.User.Edit(user)
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}

	//Delete all activation tokens for the activated user.
	err = app.Token.RemoveAllForUser(models.ScopeActivation, user.ID)
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
