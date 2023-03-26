package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/handlers"
)

func NewRouter(a app.Application) *gin.Engine {
	r := gin.Default()
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
		/*
			movies.GET("/", func(c *gin.Context) {
				handlers.listMovieHandler(c, a)
			})
			movies.PUT("/:id", func(c *gin.Context) {
				handlers.editMovieHandler(c, a)
			})
			movies.DELETE("/:id", func(c *gin.Context) {
				handlers.createMovieHandler(c, a)
			})*/
	}
	return r
}
