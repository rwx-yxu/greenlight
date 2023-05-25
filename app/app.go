package app

import (
	"database/sql"
	"net/http"

	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/jsonlog"
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

type Services struct {
	Movie services.MovieReadWriteDeleter
}

type Application struct {
	Config *Config
	Logger *jsonlog.Logger
	Services
}

func NewApp(conf Config, db *sql.DB, log *jsonlog.Logger) *Application {
	ms := services.NewMovie(brokers.NewMovie(db))
	return &Application{
		Config: &conf,
		Logger: log,
		Services: Services{
			Movie: ms,
		},
	}
}

func (app *Application) LogError(r *http.Request, err error) {
	// Use the PrintError() method to log the error message, and include the current
	// request method and URL as properties in the log entry.
	app.Logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}
