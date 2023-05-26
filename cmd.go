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
		// Create a shutdownError channel. We will use this to receive any errors returned
		// by the graceful Shutdown() function.
		shutdownError := make(chan error)

		go func() {
			// Intercept the signals, as before.
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			s := <-quit

			// Update the log entry to say "shutting down server" instead of "caught signal".
			app.Logger.PrintInfo("shutting down server", map[string]string{
				"signal": s.String(),
			})

			// Create a context with a 20-second timeout.
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// Call Shutdown() on our server, passing in the context we just made.
			// Shutdown() will return nil if the graceful shutdown was successful, or an
			// error (which may happen because of a problem closing the listeners, or
			// because the shutdown didn't complete before the 20-second context deadline is
			// hit). We relay this return value to the shutdownError channel.
			shutdownError <- srv.Shutdown(ctx)
		}()
		// Again, we use the PrintInfo() method to write a "starting server" message at the
		// INFO level. But this time we pass a map containing additional properties (the
		// operating environment and server address) as the final parameter.
		logger.PrintInfo("starting server", map[string]string{
			"addr": fmt.Sprintf(":%d", config.Server.Port),
			"env":  config.Server.Env,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
		// Otherwise, we wait to receive the return value from Shutdown() on the
		// shutdownError channel. If return value is an error, we know that there was a
		// problem with the graceful shutdown and we return the error.
		err = <-shutdownError
		if err != nil {
			return err
		}

		// At this point we know that the graceful shutdown completed successfully and we
		// log a "stopped server" message.
		app.Logger.PrintInfo("stopped server", map[string]string{
			"addr": fmt.Sprintf(":%d", config.Server.Port),
		})
		return nil

	},
}
