package greenlight

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/database"
	"github.com/rwx-yxu/greenlight/internal/services"
	"github.com/rwx-yxu/greenlight/routes"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
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
		port, err := x.Caller.C("server.port")
		if err != nil {
			return errors.New("config has not been initialized. User command 'greenlight help to set config file'")
		}
		env, err := x.Caller.C("server.env")
		if err != nil {
			return errors.New("config has not been initialized. User command 'greenlight help to set config file'")
		}
		dsn, err := x.Caller.C("db.dsn")
		if err != nil {
			return errors.New("PostgreSQL DSN not set. User command 'greenlight conf edit' add the PostreSQL DSN")
		}
		maxOpenConns, err := x.Caller.C("db.maxOpenConns")
		if err != nil {
			return errors.New("PostgreSQL max open connections not set. User command 'greenlight conf edit' add the PostreSQL max open connections.")
		}
		maxIdleConns, err := x.Caller.C("db.maxIdleConns")
		if err != nil {
			return errors.New("PostgreSQL max idle connections not set. User command 'greenlight conf edit' add the PostreSQL max    idle connections.")
		}
		maxIdleTime, err := x.Caller.C("db.maxIdleTime")
		if err != nil {
			return errors.New("PostgreSQL max open connections not set. User command 'greenlight conf edit' add the PostreSQL max    Idle time.")
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		maxOpenConnections, err := strconv.Atoi(maxOpenConns)
		if err != nil {
			return err
		}
		maxIdleConnections, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			return err
		}
		db, err := database.OpenPostgres(dsn, maxOpenConnections, maxIdleConnections, maxIdleTime)
		if err != nil {
			return err
		}
		defer db.Close()
		app := app.NewApp(p, env, x.Caller.GetVersion(), services.NewMovie())

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", app.Config.Port),
			Handler:      routes.NewRouter(*app),
			IdleTimeout:  time.Minute,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
		return nil
	},
}
