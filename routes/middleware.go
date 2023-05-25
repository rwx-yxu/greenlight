package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/handlers"
	"golang.org/x/time/rate"
)

func RateLimit(app app.Application) gin.HandlerFunc {
	// Initialize a new rate limiter which allows an average of 2 requests per second,
	// with a maximum of 4 requests in a single ‘burst’.
	limiter := rate.NewLimiter(2, 4)
	return func(c *gin.Context) {
		// Call limiter.Allow() to see if the request is permitted, and if it's not,
		// then we call the rateLimitExceededResponse() helper to return a 429 Too Many
		// Requests response (we will create this helper in a minute).
		if !limiter.Allow() {
			handlers.ErrorResponse(c, app, handlers.RateLimitExceededError())
			return
		}

		c.Next()
	}
}
