package app

import (
	"log"
	"net/http"
	"os"

	"github.com/rwx-yxu/greenlight/internal/services"
)

type Config struct {
	Port    int
	Env     string
	Version string
}

type Application struct {
	Config       *Config
	Logger       *log.Logger
	MovieService services.MovieValidator
}

func NewApp(p int, e string, v string, ms services.MovieValidator) *Application {
	return &Application{
		Config: &Config{
			Port:    p,
			Env:     e,
			Version: v,
		},
		Logger:       log.New(os.Stdout, "", log.Ldate|log.Ltime),
		MovieService: ms,
	}
}

func (app *Application) LogError(r *http.Request, err error) {
	app.Logger.Print(err)
}
