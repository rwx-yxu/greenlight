package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
	"github.com/rwx-yxu/greenlight/internal/services"
	"github.com/rwx-yxu/greenlight/internal/validator"
)

func AuthenticationTokenHandler(c *gin.Context, app app.Application) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := ReadJSON(c, &input)
	if err != nil {
		ErrorResponse(c, app, StatusBadRequestError(err))
		return
	}
	v := validator.New()
	services.ValidateEmail(v, input.Email)
	services.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		ErrorResponse(c, app, FailedValidationResponse(v.Errors))
	}
	user, err := app.User.FindByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, brokers.ErrRecordNotFound):
			ErrorResponse(c, app, InvalidCredentialsError())
		default:
			ErrorResponse(c, app, InternalServerError(err))
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		ErrorResponse(c, app, InternalServerError(err))
		return
	}
	if !match {
		ErrorResponse(c, app, InvalidCredentialsError())
	}
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
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

	c.JSON(http.StatusCreated, gin.H{"token": token})
}
