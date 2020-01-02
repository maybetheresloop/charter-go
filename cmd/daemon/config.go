package main

import (
	"github.com/maybetheresloop/charter-go"
	"github.com/urfave/cli"
)

func populateFromCliContext(conf *charter.Config, ctx *cli.Context) {
	if ctx.IsSet("no-anonymous") {
		conf.NoAnonymous = true
	}

	if ctx.IsSet("anonymous-only") {
		conf.AnonymousOnly = true
	}
}
