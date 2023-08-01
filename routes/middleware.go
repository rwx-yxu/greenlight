package routes

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/handlers"
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
	"github.com/rwx-yxu/greenlight/internal/validator"
	"golang.org/x/time/rate"
)

func RateLimit(app app.Application) gin.HandlerFunc {
	// Define a client struct to hold the rate limiter and last seen time for each
	// client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		if app.Config.Limiter.Enabled {
			ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
			if err != nil {
				handlers.ErrorResponse(c, app, handlers.InternalServerError(err))
				return
			}
			mu.Lock()
			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new rate limiter and add the IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.Config.Limiter.RPS), app.Config.Limiter.Burst),
				}
			}

			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

			// Call the Allow() method on the rate limiter for the current IP address. If
			// the request isn't allowed, unlock the mutex and send a 429 Too Many Requests
			// response, just like before.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				handlers.ErrorResponse(c, app, handlers.RateLimitExceededError())
				return
			}
			mu.Unlock()
		}
		c.Next()
	}
}

func Authenticate(app app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		c.Writer.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no such header found.
		authorizationHeader := c.GetHeader("Authorization")

		// If there is no Authorization header found, use the contextSetUser() helper
		// that we just made to add the AnonymousUser to the request context. Then we
		// call the next handler in the chain and return without executing any of the
		// code below.
		if authorizationHeader == "" {
			c.Set("user", models.AnonymousUser)
			c.Next()
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into its constituent parts, and if the
		// header isn't in the expected format we return a 401 Unauthorized response
		// using the invalidAuthenticationTokenResponse() helper (which we will create
		// in a moment).
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			handlers.ErrorResponse(c, app, handlers.InvalidAuthenticationToken(c))
			c.Abort()
			return
		}

		// Extract the actual authentication token from the header parts.
		token := headerParts[1]

		// Validate the token to make sure it is in a sensible format.
		v := validator.New()

		// If the token isn't valid, use the invalidAuthenticationToken()
		// helper to send a response, rather than the failedValidation() helper
		// that we'd normally use.
		if app.Token.ValidatePlainText(v, token); !v.Valid() {
			handlers.ErrorResponse(c, app, handlers.InvalidAuthenticationToken(c))
			c.Abort()
			return
		}

		// Retrieve the details of the user associated with the authentication token,
		// again calling the invalidAuthenticationTokenResponse() helper if no
		// matching record was found. IMPORTANT: Notice that we are using
		// ScopeAuthentication as the first parameter here.
		user, err := app.User.FindByToken(models.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, brokers.ErrRecordNotFound):
				handlers.ErrorResponse(c, app, handlers.InvalidAuthenticationToken(c))
			default:
				handlers.ErrorResponse(c, app, handlers.InternalServerError(err))
			}
			c.Abort()
			return
		}

		// Call the contextSetUser() helper to add the user information to the request
		// context.
		c.Set("user", user)

		// Call the next handler in the chain.
		c.Next()
	}
}
