package greenlight

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
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
		config.CORS.TrustedOrigins = strings.Fields(config.CORS.Origins)
		logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
		db, err := database.OpenPostgres(config)
		if err != nil {
			return err
		}
		logger.PrintInfo("database connection pool established", nil)
		defer db.Close()
		config.Server.Version = x.Caller.GetVersion()
		expvar.NewString("version").Set(config.Server.Version)
		// Publish the number of active goroutines.
		expvar.Publish("goroutines", expvar.Func(func() any {
			return runtime.NumGoroutine()
		}))

		// Publish the database connection pool statistics.
		expvar.Publish("database", expvar.Func(func() any {
			return db.Stats()
		}))

		// Publish the current Unix timestamp.
		expvar.Publish("timestamp", expvar.Func(func() any {
			return time.Now().Unix()
		}))
		app := app.NewApp(config, db, logger)

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Server.Port),
			Handler:      routes.NewRouter(*app),
			ErrorLog:     log.New(logger, "", 0),
			IdleTimeout:  time.Minute,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		shutdownError := make(chan error)
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			s := <-quit
			app.Logger.PrintInfo("caught signal", map[string]string{
				"signal": s.String(),
			})

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call Shutdown() on the server like before, but now we only send on the
			// shutdownError channel if it returns an error.
			err := srv.Shutdown(ctx)
			if err != nil {
				shutdownError <- err
			}
			// Log a message to say that we're waiting for any background goroutines to
			// complete their tasks.
			app.Logger.PrintInfo("completing background tasks", map[string]string{
				"addr": srv.Addr,
			})

			// Call Wait() to block until our WaitGroup counter is zero --- essentially
			// blocking until the background goroutines have finished. Then we return nil on
			// the shutdownError channel, to indicate that the shutdown completed without
			// any issues.
			app.WG.Wait()
			shutdownError <- nil
		}()
		app.Logger.PrintInfo("starting server", map[string]string{
			"addr": srv.Addr,
			"env":  app.Config.Server.Env,
		})

		err = srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		err = <-shutdownError
		if err != nil {
			return err
		}

		app.Logger.PrintInfo("stopped server", map[string]string{
			"addr": srv.Addr,
		})
		return nil

	},
}
