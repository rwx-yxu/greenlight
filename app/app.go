package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/rwx-yxu/greenlight/internal/brokers"
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
	Logger *log.Logger
	Services
}

func NewApp(conf Config, db *sql.DB) *Application {
	ms := services.NewMovie(brokers.NewMovie(db))
	return &Application{
		Config: &conf,
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		Services: Services{
			Movie: ms,
		},
	}
}

func (app *Application) LogError(r *http.Request, err error) {
	app.Logger.Print(err)
}
