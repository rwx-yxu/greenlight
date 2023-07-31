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
	r.Use(gin.Recovery())
	r.Use(RateLimit(a))
	v1 := r.Group("/v1")
	v1.GET("/healthcheck", func(c *gin.Context) {
		handlers.HealthcheckHandler(c, a)
	})
	movies := v1.Group("/movies")
	{
		movies.POST("", func(c *gin.Context) {
			handlers.CreateMovieHandler(c, a)
		})
		movies.GET("/:id", func(c *gin.Context) {
			handlers.ShowMovieHandler(c, a)
		})
		movies.PATCH("/:id", func(c *gin.Context) {
			handlers.UpdateMovieHandler(c, a)
		})
		movies.DELETE("/:id", func(c *gin.Context) {
			handlers.DeleteMovieHandler(c, a)
		})
		movies.GET("", func(c *gin.Context) {
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
