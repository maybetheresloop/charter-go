package main

import (
	"os"

	"github.com/BurntSushi/toml"

	"github.com/maybetheresloop/charter-go"
	"github.com/urfave/cli"
)

func defaultConfig() *charter.Config {
	defaultDir, err := os.UserHomeDir()
	if err != nil {
		defaultDir = "/"
	}

	return &charter.Config{
		Addr:       ":5678",
		DefaultDir: defaultDir,
		PassivePortRange: charter.PassivePortRange{
			From: 40001,
			To:   40009,
		},
	}
}

func populateFromCliContext(conf *charter.Config, ctx *cli.Context) {
	if ctx.IsSet("no-anonymous") {
		conf.NoAnonymous = true
	}

	if ctx.IsSet("anonymous-only") {
		conf.AnonymousOnly = true
	}
}

func populateFromFile(conf *charter.Config, filename string) error {
	_, err := toml.DecodeFile(filename, conf)
	return err
}
