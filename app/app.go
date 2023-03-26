package app

import (
	"log"
	"os"
)

type Config struct {
	Port int
	Env  string
}

type Application struct {
	Config *Config
	Logger *log.Logger
}

func NewApp(p int, e string) *Application {
	return &Application{
		Config: &Config{
			Port: p,
			Env:  e,
		},
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}
