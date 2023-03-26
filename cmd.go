package greenlight

import (
	"fmt"

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
		fmt.Println("Hello world")
		return nil
	},
}
