package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/handlers"
)

func NewRouter(a app.Application) *gin.Engine {
	r := gin.Default()
	r.NoMethod(MethodNotAllowed(a))
	r.NoRoute(NotFound(a))
	r.Use(gin.Recovery(), CORS(a), RateLimit(a), Authenticate(a))
	v1 := r.Group("/v1")
	v1.GET("/healthcheck", func(c *gin.Context) {
		handlers.HealthcheckHandler(c, a)
	})
	movies := v1.Group("/movies")
	movies.Use(RequireActivated(a))
	{
		movies.POST("", RequirePermission(a, "movies:write"), func(c *gin.Context) {
			handlers.CreateMovieHandler(c, a)
		})
		movies.GET("/:id", RequirePermission(a, "movies:read"), func(c *gin.Context) {
			handlers.ShowMovieHandler(c, a)
		})
		movies.PATCH("/:id", RequirePermission(a, "movies:write"), func(c *gin.Context) {
			handlers.UpdateMovieHandler(c, a)
		})
		movies.DELETE("/:id", RequirePermission(a, "movies:write"), func(c *gin.Context) {
			handlers.DeleteMovieHandler(c, a)
		})
		movies.GET("", RequirePermission(a, "movies:read"), func(c *gin.Context) {
			handlers.ListMoviesHandler(c, a)
		})
	}
	users := v1.Group("/users")
	{
		users.POST("", func(c *gin.Context) {
			handlers.RegisterUserHandler(c, a)
		})
		users.PUT("/activated", func(c *gin.Context) {
			handlers.ActivateUserHandler(c, a)
		})
	}
	tokens := v1.Group("/tokens")
	{
		tokens.POST("/authentication", func(c *gin.Context) {
			handlers.AuthenticationTokenHandler(c, a)
		})
	}
	return r
}
