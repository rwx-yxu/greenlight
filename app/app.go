package app

import (
	"log"
	"net/http"
	"os"
)

type Config struct {
	Port    int
	Env     string
	Version string
}

type Application struct {
	Config *Config
	Logger *log.Logger
}

func NewApp(p int, e string, v string) *Application {
	return &Application{
		Config: &Config{
			Port:    p,
			Env:     e,
			Version: v,
		},
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

func (app *Application) LogError(r *http.Request, err error) {
	app.Logger.Print(err)
}
