package greenlight

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rwx-yxu/greenlight/app"
	"github.com/rwx-yxu/greenlight/database"
	"github.com/rwx-yxu/greenlight/internal/services"
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

		db, err := database.OpenPostgres(config)
		if err != nil {
			return err
		}
		defer db.Close()
		config.Server.Version = x.Caller.GetVersion()
		app := app.NewApp(config, services.NewMovie())

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Server.Port),
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
