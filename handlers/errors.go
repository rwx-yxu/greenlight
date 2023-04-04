package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponseBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type HandleError struct {
	StatusCode int
	Response   ErrorResponseBody
}

var HttpErrorMessages = map[int]string{
	http.StatusNotFound:            "the requested resource could not be found",
	http.StatusBadRequest:          "the requested action cannot be performed with the provided parameters",
	http.StatusInternalServerError: "Internal server error",
	http.StatusMethodNotAllowed:    "the %s method is not supported for this resource",
	// Add more status codes and messages as needed
}

var HttpErrorCodeStrings = map[int]string{
	http.StatusNotFound:         "NOT_FOUND",
	http.StatusBadRequest:       "BAD_REQUEST",
	http.StatusMethodNotAllowed: "METHOD_NOT_ALLOWED",
}

func (h HandleError) Error() string {
	return fmt.Sprintf("Status Code: %d, Response: %v", h.StatusCode, h.Response)
}

func ErrorResponse(c *gin.Context, app app.Application, err error) {
	var handleError HandleError
	if errors.As(err, &handleError) {
		app.LogError(c.Request, err)
		c.JSON(handleError.StatusCode, gin.H{"error": handleError.Response})
		return
	}
	app.LogError(c.Request, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func NotFoundError(origErr error) error {
	details := []ErrorDetail{}

	if origErr != nil {
		details = append(details, ErrorDetail{"original_error", origErr.Error()})
	}

	response := ErrorResponseBody{
		Code:    HttpErrorCodeStrings[http.StatusNotFound],
		Message: HttpErrorMessages[http.StatusNotFound],
		Details: details,
	}

	return fmt.Errorf("%w", HandleError{
		StatusCode: http.StatusNotFound,
		Response:   response,
	})
}

func NotAllowedError(origErr error, method string) error {
	details := []ErrorDetail{}

	if origErr != nil {
		details = append(details, ErrorDetail{"original_error", origErr.Error()})
	}

	response := ErrorResponseBody{
		Code:    HttpErrorCodeStrings[http.StatusMethodNotAllowed],
		Message: HttpErrorMessages[http.StatusMethodNotAllowed],
		Details: details,
	}

	return fmt.Errorf("%w", HandleError{
		StatusCode: http.StatusMethodNotAllowed,
		Response:   response,
	})
}

func StatusBadRequestError(origErr error) error {
	details := []ErrorDetail{}

	if origErr != nil {
		details = append(details, ErrorDetail{"original_error", origErr.Error()})
	}

	response := ErrorResponseBody{
		Code:    HttpErrorCodeStrings[http.StatusBadRequest],
		Message: HttpErrorMessages[http.StatusBadRequest],
		Details: details,
	}

	return fmt.Errorf("%w", HandleError{
		StatusCode: http.StatusBadRequest,
		Response:   response,
	})
}

func TriageJSONError(err error) error {
	switch e := err.(type) {
	case *json.SyntaxError:
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", e.Offset)
	case *json.UnmarshalTypeError:
		if e.Field != "" {
			return fmt.Errorf("body contains incorrect JSON type for field %q", e.Field)
		}
		return fmt.Errorf("body contains incorrect JSON type (at character %d)", e.Offset)
	case *json.InvalidUnmarshalError:
		panic(e)
	case *http.MaxBytesError:
		return fmt.Errorf("body must not be larger than %d bytes", e.Limit)
	default:
		if err == io.EOF {
			return errors.New("body must not be empty")
		}
		return e
	}
}
