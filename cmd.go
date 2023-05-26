package greenlight

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/database"
	"github.com/rwx-yxu/greenlight/internal/jsonlog"
	"github.com/rwx-yxu/greenlight/routes"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"gopkg.in/yaml.v3"
)

func init() {
	Z.Conf.SoftInit()
}

var Cmd = &Z.Cmd{
	Name:      `greenlight`,
	Version:   `v0.0.1`,
	Copyright: `Copyright 2023 Yongle Xu`,
	License:   `Apache-2.0`,
	Site:      `yonglexu.dev`,
	Source:    `git@github.com:rwx-yxu/greenlight.git`,
	Issues:    `github.com/rwx-yxu/greenlight/issues`,

	Commands: []*Z.Cmd{
		StartCmd,

		// standard external branch imports (see rwxrob/{help,conf,vars})
		help.Cmd, conf.Cmd,
	},
	Summary:     help.S(_greenlight),
	Description: help.D(_greenlight),
}

var StartCmd = &Z.Cmd{
	Name:        `start`,
	Aliases:     []string{`s`},
	Commands:    []*Z.Cmd{help.Cmd},
	Summary:     help.S(_start),
	Description: help.D(_start),
	Call: func(x *Z.Cmd, _ ...string) error {
		var config app.Config
		c, err := x.Caller.C("")
		if err != nil {
			return errors.New("Config has not been initialised")
		}
		err = yaml.Unmarshal([]byte(c), &config)
		if err != nil {
			return err
		}
		logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
		db, err := database.OpenPostgres(config)
		if err != nil {
			return err
		}
		logger.PrintInfo("database connection pool established", nil)
		defer db.Close()
		config.Server.Version = x.Caller.GetVersion()

		app := app.NewApp(config, db, logger)
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Server.Port),
			Handler:      routes.NewRouter(*app),
			ErrorLog:     log.New(logger, "", 0),
			IdleTimeout:  time.Minute,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		go func() {
			// service connections
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				app.Logger.PrintInfo("starting server", map[string]string{
					"addr": fmt.Sprintf(":%d", config.Server.Port),
					"env":  app.Config.Server.Env,
				})
			}
		}()
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		quit := make(chan os.Signal)
		// kill (no param) default send syscanll.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		app.Logger.PrintInfo("shutting down server", map[string]string{})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			app.Logger.PrintFatal(err, map[string]string{
				"addr": fmt.Sprintf(":%d", config.Server.Port),
			})
		}
		// catching ctx.Done(). timeout of 5 seconds.
		select {
		case <-ctx.Done():
			app.Logger.PrintInfo("time out of 5 seconds", map[string]string{})

		}
		app.Logger.PrintInfo("stopped server", map[string]string{})

		return nil

	},
}
