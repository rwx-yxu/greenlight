package app

import (
	"log"
	"net/http"
	"os"

	"github.com/rwx-yxu/greenlight/internal/services"
)

/*
type Config struct {
	Port    int
	Env     string
	Version string
}*/
type Config struct {
	Server struct {
		Port    int    `yaml:"port"`
		Env     string `yaml:"env"`
		Version string
	} `yaml:"server"`
	DB struct {
		DSN          string `yaml:"dsn"`
		MaxOpenConns int    `yaml:"maxOpenConns"`
		MaxIdleConns int    `yaml:"maxIdleConns"`
		MaxIdleTime  string `yaml:"maxIdleTime"`
	} `yaml:"db"`
}
type Application struct {
	Config       *Config
	Logger       *log.Logger
	MovieService services.MovieValidator
}

func NewApp(conf Config, ms services.MovieValidator) *Application {
	return &Application{
		Config:       &conf,
		Logger:       log.New(os.Stdout, "", log.Ldate|log.Ltime),
		MovieService: ms,
	}
}

func (app *Application) LogError(r *http.Request, err error) {
	app.Logger.Print(err)
}
