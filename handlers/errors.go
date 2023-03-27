package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
)

type HandleError struct {
	StatusCode int
	error
}

var httpErrorMessages = map[int]string{
	http.StatusNotFound:            "the requested resource could not be found",
	http.StatusBadRequest:          "Bad request",
	http.StatusInternalServerError: "Internal server error",
	http.StatusMethodNotAllowed:    "the %s method is not supported for this resource",
	// Add more status codes and messages as needed
}

func ErrorResponse(c *gin.Context, app app.Application, err error) {
	var handleError HandleError
	if errors.As(err, &handleError) {
		app.LogError(c.Request, err)
		c.JSON(handleError.StatusCode, gin.H{"error": handleError.Error()})
		return
	}
	app.LogError(c.Request, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func NotFoundError(origErr error) HandleError {
	msg := httpErrorMessages[http.StatusNotFound]
	if origErr != nil {
		msg = fmt.Sprintf("%s: %w", msg, origErr)
	}
	return HandleError{
		error:      errors.New(msg),
		StatusCode: http.StatusNotFound,
	}
}

func NotAllowedError(origErr error, method string) HandleError {
	msg := fmt.Sprintf(httpErrorMessages[http.StatusMethodNotAllowed], method)
	if origErr != nil {
		msg = fmt.Sprintf("%s: %w", msg, origErr)
	}
	return HandleError{
		error:      errors.New(msg),
		StatusCode: http.StatusMethodNotAllowed,
	}
}
