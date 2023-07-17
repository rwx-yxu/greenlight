package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/jsonlog"
	"github.com/rwx-yxu/greenlight/internal/mailer"
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
	Limiter struct {
		RPS     float64 `yaml:"rps"`
		Burst   int     `yaml:"burst"`
		Enabled bool    `yaml:"enabled"`
	} `yaml:limiter`
	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Sender   string `yaml:"sender"`
	} `yaml:"smtp"`
}

type Services struct {
	Movie services.MovieReadWriteDeleter
	User  services.UserWriter
}

type Application struct {
	Config *Config
	Logger *jsonlog.Logger
	Services
	SMTP mailer.Mailer
}

func NewApp(conf Config, db *sql.DB, log *jsonlog.Logger) *Application {
	ms := services.NewMovie(brokers.NewMovie(db))
	us := services.NewUser(brokers.NewUser(db))
	return &Application{
		Config: &conf,
		Logger: log,
		Services: Services{
			Movie: ms,
			User:  us,
		},
		SMTP: mailer.New(conf.SMTP.Host, conf.SMTP.Port, conf.SMTP.Username, conf.SMTP.Password, conf.SMTP.Sender),
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

//Background is a function which will execute an accepted function as a go routine
//The go routine will ensure that there is a deferred and recover.
func (app *Application) Background(fn func()) {
	// Launch a background goroutine.
	go func() {
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				app.Logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}
